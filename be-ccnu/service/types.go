package service

import (
	"context"
	"github.com/asynccnu/be-ccnu/domain"
	"github.com/asynccnu/be-ccnu/pkg/logger"
	"time"
)

type CCNUService interface {
	Login(ctx context.Context, studentId string, password string) (bool, error)
	GetSelfCourseList(ctx context.Context, studentId, password, year, term string) ([]domain.Course, error)
	// GetSelfGradeList 这个是只能获取总分，没有聚合平时成绩等细节，现在主要用于准确获取个人历史课程
	GetSelfGradeList(ctx context.Context, studentId, password, year, term string) ([]domain.Grade, error)
	// GetAllDetailOfGrade 获取所有成绩的所有细节
	GetDetailOfGradeList(ctx context.Context, studentId string, password string, year string, term string) ([]domain.Grade, error)
}

type ccnuService struct {
	timeout time.Duration
	l       logger.Logger
}

func NewCCNUService(l logger.Logger) CCNUService {
	return &ccnuService{
		timeout: time.Second * 5,
		l:       l,
	}
}
