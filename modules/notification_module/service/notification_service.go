package service

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"auto_park/modules/notification_module/dto"
	"auto_park/modules/notification_module/models"
	"auto_park/modules/notification_module/repository"
	"auto_park/modules/notification_module/websocket"
)

const (
	NotificationTypePartRequestCreated  = "part_request_created"
	NotificationTypePartRequestApproved = "part_request_approved"
	NotificationTypePartRequestRejected = "part_request_rejected"

	NotificationTypeVehiclePartReplacement7Days = "vehicle_part_replacement_7_days"
	NotificationTypeVehiclePartReplacementToday = "vehicle_part_replacement_today"

	WarehouseManagerRoleCode       = "warehouse_manager"
	WarehouseManagerFallbackRoleID = int64(5)
)

type NotificationService interface {
	CreateForUser(ctx context.Context, userID int64, typeCode string, title string, message string, contextData map[string]any) (*dto.NotificationResponse, error)
	CreateForUserOnce(ctx context.Context, userID int64, dedupKey string, typeCode string, title string, message string, contextData map[string]any) (*dto.NotificationResponse, error)
	CreateForUsers(ctx context.Context, userIDs []int64, typeCode string, title string, message string, contextData map[string]any) ([]dto.NotificationResponse, error)
	CreateForUsersOnce(ctx context.Context, userIDs []int64, dedupKey string, typeCode string, title string, message string, contextData map[string]any) ([]dto.NotificationResponse, error)
	CreateForRole(ctx context.Context, roleCode string, fallbackRoleID int64, typeCode string, title string, message string, contextData map[string]any) ([]dto.NotificationResponse, error)
	CreateForRoleOnce(ctx context.Context, roleCode string, fallbackRoleID int64, dedupKey string, typeCode string, title string, message string, contextData map[string]any) ([]dto.NotificationResponse, error)
	List(ctx context.Context, userID int64, onlyUnread bool, limit int, offset int) (*dto.NotificationListResponse, error)
	CountUnread(ctx context.Context, userID int64) (int64, error)
	MarkAsRead(ctx context.Context, userID int64, notificationID int64) (bool, error)
	MarkAllAsRead(ctx context.Context, userID int64) (int64, error)
	PushUnreadSnapshot(ctx context.Context, userID int64) error
}

type notificationService struct {
	repo repository.NotificationRepository
	hub  *websocket.Hub
}

func NewNotificationService(repo repository.NotificationRepository, hub *websocket.Hub) NotificationService {
	return &notificationService{repo: repo, hub: hub}
}

func (s *notificationService) CreateForUser(ctx context.Context, userID int64, typeCode string, title string, message string, contextData map[string]any) (*dto.NotificationResponse, error) {
	return s.createForUser(ctx, userID, nil, typeCode, title, message, contextData)
}

func (s *notificationService) CreateForUserOnce(ctx context.Context, userID int64, dedupKey string, typeCode string, title string, message string, contextData map[string]any) (*dto.NotificationResponse, error) {
	dedupKey = strings.TrimSpace(dedupKey)
	if dedupKey == "" {
		return nil, fmt.Errorf("notification dedup_key is required")
	}
	return s.createForUser(ctx, userID, &dedupKey, typeCode, title, message, contextData)
}

func (s *notificationService) createForUser(ctx context.Context, userID int64, dedupKey *string, typeCode string, title string, message string, contextData map[string]any) (*dto.NotificationResponse, error) {
	if userID <= 0 {
		return nil, fmt.Errorf("invalid notification user_id")
	}
	typeCode = strings.TrimSpace(typeCode)
	if typeCode == "" {
		return nil, fmt.Errorf("notification type_code is required")
	}
	title = strings.TrimSpace(title)
	if title == "" {
		return nil, fmt.Errorf("notification title is required")
	}
	message = strings.TrimSpace(message)
	if message == "" {
		return nil, fmt.Errorf("notification message is required")
	}
	if contextData == nil {
		contextData = map[string]any{}
	}

	item, err := s.repo.Create(ctx, repository.CreateNotificationParams{
		UserID:   userID,
		TypeCode: typeCode,
		Title:    title,
		Message:  message,
		Context:  contextData,
		DedupKey: dedupKey,
	})
	if err != nil {
		return nil, err
	}
	if item == nil {
		return nil, nil
	}

	resp := mapNotificationToDTO(*item)
	s.pushToUser(userID, "notification.created", resp)

	count, err := s.repo.CountUnread(ctx, userID)
	if err == nil {
		s.pushToUser(userID, "notification.unread_count", dto.NotificationUnreadCountResponse{Count: count})
	}

	return &resp, nil
}

func (s *notificationService) CreateForUsers(ctx context.Context, userIDs []int64, typeCode string, title string, message string, contextData map[string]any) ([]dto.NotificationResponse, error) {
	return s.createForUsers(ctx, userIDs, nil, typeCode, title, message, contextData)
}

func (s *notificationService) CreateForUsersOnce(ctx context.Context, userIDs []int64, dedupKey string, typeCode string, title string, message string, contextData map[string]any) ([]dto.NotificationResponse, error) {
	dedupKey = strings.TrimSpace(dedupKey)
	if dedupKey == "" {
		return nil, fmt.Errorf("notification dedup_key is required")
	}
	return s.createForUsers(ctx, userIDs, &dedupKey, typeCode, title, message, contextData)
}

func (s *notificationService) createForUsers(ctx context.Context, userIDs []int64, dedupKey *string, typeCode string, title string, message string, contextData map[string]any) ([]dto.NotificationResponse, error) {
	unique := make(map[int64]struct{}, len(userIDs))
	items := make([]dto.NotificationResponse, 0, len(userIDs))

	for _, userID := range userIDs {
		if userID <= 0 {
			continue
		}
		if _, exists := unique[userID]; exists {
			continue
		}
		unique[userID] = struct{}{}

		item, err := s.createForUser(ctx, userID, dedupKey, typeCode, title, message, contextData)
		if err != nil {
			return items, err
		}
		if item != nil {
			items = append(items, *item)
		}
	}

	return items, nil
}

func (s *notificationService) CreateForRole(ctx context.Context, roleCode string, fallbackRoleID int64, typeCode string, title string, message string, contextData map[string]any) ([]dto.NotificationResponse, error) {
	userIDs, err := s.repo.GetUserIDsByRole(ctx, roleCode, fallbackRoleID)
	if err != nil {
		return nil, err
	}
	return s.CreateForUsers(ctx, userIDs, typeCode, title, message, contextData)
}

func (s *notificationService) CreateForRoleOnce(ctx context.Context, roleCode string, fallbackRoleID int64, dedupKey string, typeCode string, title string, message string, contextData map[string]any) ([]dto.NotificationResponse, error) {
	userIDs, err := s.repo.GetUserIDsByRole(ctx, roleCode, fallbackRoleID)
	if err != nil {
		return nil, err
	}
	return s.CreateForUsersOnce(ctx, userIDs, dedupKey, typeCode, title, message, contextData)
}

func (s *notificationService) List(ctx context.Context, userID int64, onlyUnread bool, limit int, offset int) (*dto.NotificationListResponse, error) {
	if userID <= 0 {
		return nil, fmt.Errorf("invalid user_id")
	}

	items, total, err := s.repo.ListByUser(ctx, repository.ListNotificationsParams{
		UserID:     userID,
		OnlyUnread: onlyUnread,
		Limit:      limit,
		Offset:     offset,
	})
	if err != nil {
		return nil, err
	}

	out := make([]dto.NotificationResponse, 0, len(items))
	for _, item := range items {
		out = append(out, mapNotificationToDTO(item))
	}
	unreadCount, err := s.repo.CountUnread(ctx, userID)
	if err != nil {
		return nil, err
	}

	if limit <= 0 || limit > 200 {
		limit = 50
	}
	if offset < 0 {
		offset = 0
	}

	return &dto.NotificationListResponse{
		Items:       out,
		UnreadCount: unreadCount,
		Total:       total,
		Limit:       limit,
		Offset:      offset,
	}, nil
}

func (s *notificationService) CountUnread(ctx context.Context, userID int64) (int64, error) {
	if userID <= 0 {
		return 0, fmt.Errorf("invalid user_id")
	}
	return s.repo.CountUnread(ctx, userID)
}

func (s *notificationService) MarkAsRead(ctx context.Context, userID int64, notificationID int64) (bool, error) {
	if userID <= 0 {
		return false, fmt.Errorf("invalid user_id")
	}
	if notificationID <= 0 {
		return false, fmt.Errorf("invalid notification id")
	}
	updated, err := s.repo.MarkAsRead(ctx, userID, notificationID)
	if err != nil {
		return false, err
	}
	if updated {
		if count, err := s.repo.CountUnread(ctx, userID); err == nil {
			s.pushToUser(userID, "notification.unread_count", dto.NotificationUnreadCountResponse{Count: count})
		}
	}
	return updated, nil
}

func (s *notificationService) MarkAllAsRead(ctx context.Context, userID int64) (int64, error) {
	if userID <= 0 {
		return 0, fmt.Errorf("invalid user_id")
	}
	updated, err := s.repo.MarkAllAsRead(ctx, userID)
	if err != nil {
		return 0, err
	}
	s.pushToUser(userID, "notification.unread_count", dto.NotificationUnreadCountResponse{Count: 0})
	return updated, nil
}

func (s *notificationService) PushUnreadSnapshot(ctx context.Context, userID int64) error {
	if userID <= 0 {
		return fmt.Errorf("invalid user_id")
	}
	list, err := s.List(ctx, userID, true, 50, 0)
	if err != nil {
		return err
	}
	count, err := s.repo.CountUnread(ctx, userID)
	if err != nil {
		return err
	}
	s.pushToUser(userID, "notification.unread_snapshot", list)
	s.pushToUser(userID, "notification.unread_count", dto.NotificationUnreadCountResponse{Count: count})
	return nil
}

func (s *notificationService) pushToUser(userID int64, event string, data any) {
	if s.hub == nil {
		return
	}
	payload, err := json.Marshal(dto.WebSocketMessage{Event: event, Data: data})
	if err != nil {
		return
	}
	s.hub.SendToUser(userID, payload)
}

func mapNotificationToDTO(item models.Notification) dto.NotificationResponse {
	contextData := map[string]any{}
	if len(item.Context) > 0 {
		_ = json.Unmarshal(item.Context, &contextData)
	}

	return dto.NotificationResponse{
		ID:                 item.ID,
		UserID:             item.UserID,
		NotificationTypeID: item.NotificationTypeID,
		TypeCode:           item.TypeCode,
		TypeName:           item.TypeName,
		Title:              item.Title,
		Message:            item.Message,
		Context:            contextData,
		DedupKey:           item.DedupKey,
		IsReaded:           item.IsReaded,
		ReadAt:             item.ReadAt,
		CreatedAt:          item.CreatedAt,
		UpdatedAt:          item.UpdatedAt,
	}
}
