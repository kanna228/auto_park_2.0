package service

import (
	"context"
	"fmt"
	"mime/multipart"

	"auto_park/vehicle_module/dto"
)

func (s *vehicleService) UploadPhoto(ctx context.Context, id int64, fileHeader any) (*dto.VehicleResponse, error) {
	fh, ok := fileHeader.(*multipart.FileHeader)
	if !ok || fh == nil {
		return nil, fmt.Errorf("photo file is required")
	}

	vehicle, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if vehicle == nil {
		return nil, nil
	}

	relativePath, err := s.storage.Save(id, fh, vehicle.PhotoPath)
	if err != nil {
		return nil, err
	}

	ok, err = s.repo.UpdatePhotoPath(ctx, id, relativePath)
	if err != nil {
		return nil, err
	}
	if !ok {
		return nil, nil
	}

	updated, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if updated == nil {
		return nil, nil
	}

	resp := mapVehicleToDTO(*updated)
	return &resp, nil
}

func (s *vehicleService) DeletePhoto(ctx context.Context, id int64) (*dto.VehicleResponse, error) {
	vehicle, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if vehicle == nil {
		return nil, nil
	}

	if vehicle.PhotoPath != "" {
		if err := s.storage.Delete(vehicle.PhotoPath); err != nil {
			return nil, err
		}
	}

	ok, err := s.repo.UpdatePhotoPath(ctx, id, "")
	if err != nil {
		return nil, err
	}
	if !ok {
		return nil, nil
	}

	updated, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if updated == nil {
		return nil, nil
	}

	resp := mapVehicleToDTO(*updated)
	return &resp, nil
}
