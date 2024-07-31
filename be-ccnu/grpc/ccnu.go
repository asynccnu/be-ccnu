package grpc

import (
	"context"
	ccnuv1 "github.com/asynccnu/be-api/gen/proto/ccnu/v1"
	"github.com/asynccnu/be-ccnu/domain"
	"github.com/asynccnu/be-ccnu/service"
	"github.com/ecodeclub/ekit/slice"
	"google.golang.org/grpc"
)

type CCNUServiceServer struct {
	ccnuv1.UnimplementedCCNUServiceServer
	ccnu service.CCNUService
}

func NewCCNUServiceServer(ccnu service.CCNUService) *CCNUServiceServer {
	return &CCNUServiceServer{ccnu: ccnu}
}

func (s *CCNUServiceServer) Register(server grpc.ServiceRegistrar) {
	ccnuv1.RegisterCCNUServiceServer(server, s)
}

func (s *CCNUServiceServer) Login(ctx context.Context, request *ccnuv1.LoginRequest) (*ccnuv1.LoginResponse, error) {
	success, err := s.ccnu.Login(ctx, request.GetStudentId(), request.GetPassword())
	return &ccnuv1.LoginResponse{Success: success}, err
}

func (s *CCNUServiceServer) CourseList(ctx context.Context, request *ccnuv1.CourseListRequest) (*ccnuv1.CourseListResponse, error) {
	var courseVos []*ccnuv1.Course
	// 利用成绩接口，或者老接口
	if request.GetSource() == ccnuv1.Source_GradeApi {
		grades, err := s.ccnu.GetSelfGradeList(ctx, request.GetStudentId(), request.GetPassword(),
			request.GetYear(), request.GetTerm())
		if err != nil {
			return nil, err
		}
		courseVos = slice.Map(grades, func(idx int, src domain.Grade) *ccnuv1.Course {
			courseV := convertToCourseV(src.Course)
			courseV.Year = src.Year
			courseV.Term = src.Term
			return courseV
		})
	} else {
		courses, err := s.ccnu.GetSelfCourseList(ctx, request.GetStudentId(), request.GetPassword(),
			request.GetYear(), request.GetTerm())
		if err != nil {
			return nil, err
		}
		courseVos = slice.Map(courses, func(idx int, src domain.Course) *ccnuv1.Course {
			return convertToCourseV(src)
		})
	}
	return &ccnuv1.CourseListResponse{
		Courses: courseVos,
	}, nil
}

// todo: 修改Grade服务，废弃该方法，使用GetGrades代替
func (s *CCNUServiceServer) GetAllGrades(ctx context.Context, request *ccnuv1.GetAllGradesRequest) (*ccnuv1.GetAllGradesResponse, error) {
	grades, err := s.ccnu.GetDetailOfGradeList(ctx, request.GetStudentId(), request.GetPassword(), "", "")
	return &ccnuv1.GetAllGradesResponse{
		Grades: slice.Map(grades, func(idx int, src domain.Grade) *ccnuv1.Grade {
			return convertToGradeV(src)
		}),
	}, err
}

func (s *CCNUServiceServer) GetGrades(ctx context.Context, request *ccnuv1.GetGradesRequest) (*ccnuv1.GetGradesResponse, error) {
	grades, err := s.ccnu.GetDetailOfGradeList(ctx, request.GetStudentId(), request.GetPassword(), request.GetYear(), request.GetYear())
	return &ccnuv1.GetGradesResponse{
		Grades: slice.Map(grades, func(idx int, src domain.Grade) *ccnuv1.Grade {
			return convertToGradeV(src)
		}),
	}, err
}

func convertToCourseV(c domain.Course) *ccnuv1.Course {
	return &ccnuv1.Course{
		CourseCode: c.CourseId,
		Name:       c.Name,
		Teacher:    c.Teacher,
		Class:      c.Class,
		School:     c.School,
		Property:   c.Property,
		Credit:     c.Credit,
		Year:       c.Year,
		Term:       c.Term,
	}
}

func convertToGradeV(g domain.Grade) *ccnuv1.Grade {
	return &ccnuv1.Grade{
		CourseCode:    g.Course.CourseId,
		CourseName:    g.Course.Name,
		CourseTeacher: g.Course.Teacher,
		Regular:       g.Regular,
		Final:         g.Final,
		Total:         g.Total,
		Year:          g.Year,
		Term:          g.Term,
	}
}
