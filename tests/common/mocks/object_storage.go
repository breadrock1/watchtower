package mocks

import (
	"net/url"

	"github.com/stretchr/testify/mock"

	"watchtower/internal/core/cloud/domain"
	"watchtower/internal/shared/kernel"
)

type MockObjectStorage struct {
	mock.Mock
}

func (m *MockObjectStorage) GetAllBuckets(_ kernel.Ctx) ([]domain.Bucket, error) {
	args := m.Called()
	return args.Get(0).([]domain.Bucket), args.Error(1)
}

func (m *MockObjectStorage) IsBucketExist(_ kernel.Ctx, bucketID kernel.BucketID) (bool, error) {
	args := m.Called(bucketID)
	return args.Bool(0), args.Error(1)
}

func (m *MockObjectStorage) CreateBucket(_ kernel.Ctx, bucketID kernel.BucketID) error {
	args := m.Called(bucketID)
	return args.Error(0)
}

func (m *MockObjectStorage) DeleteBucket(_ kernel.Ctx, bucketID kernel.BucketID) error {
	args := m.Called(bucketID)
	return args.Error(0)
}

func (m *MockObjectStorage) GetObjectInfo(
	_ kernel.Ctx,
	bucketID kernel.BucketID,
	objID kernel.ObjectID,
) (domain.Object, error) {
	args := m.Called(bucketID, objID)
	return args.Get(0).(domain.Object), args.Error(1)
}

func (m *MockObjectStorage) GetObjectData(
	_ kernel.Ctx,
	bucketID kernel.BucketID,
	objID kernel.ObjectID,
) (domain.ObjectData, error) {
	args := m.Called(bucketID, objID)
	return args.Get(0).(domain.ObjectData), args.Error(1)
}

func (m *MockObjectStorage) StoreObject(
	_ kernel.Ctx,
	bucketID kernel.BucketID,
	params *domain.UploadObjectParams,
) (kernel.ObjectID, error) {
	args := m.Called(bucketID, params)
	return args.Get(0).(kernel.ObjectID), args.Error(1)
}

func (m *MockObjectStorage) CopyObject(_ kernel.Ctx, bucketID kernel.BucketID, params *domain.CopyObjectParams) error {
	args := m.Called(bucketID, params)
	return args.Error(0)
}

func (m *MockObjectStorage) DeleteObject(_ kernel.Ctx, bucketID kernel.BucketID, objID kernel.ObjectID) error {
	args := m.Called(bucketID, objID)
	return args.Error(0)
}

func (m *MockObjectStorage) DeleteObjects(ctx kernel.Ctx, bucketID kernel.BucketID, prefix string) error {
	args := m.Called(bucketID, prefix)
	return args.Error(0)
}

func (m *MockObjectStorage) GetBucketObjects(
	_ kernel.Ctx,
	bucketID kernel.BucketID,
	params *domain.GetObjectsParams,
) ([]domain.Object, error) {
	args := m.Called(bucketID, params)
	return args.Get(0).([]domain.Object), args.Error(1)
}

func (m *MockObjectStorage) GenShareURL(
	_ kernel.Ctx,
	bucketID kernel.BucketID,
	params *domain.ShareObjectParams,
) (*url.URL, error) {
	args := m.Called(bucketID, params)
	return args.Get(0).(*url.URL), args.Error(1)
}
