package service

import (
	"github.com/hlfshell/coppermind/internal/store"
	"github.com/hlfshell/coppermind/pkg/memory"
)

type SummaryService struct {
	db store.Store
}

func NewSummaryService(db store.Store) *SummaryService {
	return &SummaryService{
		db: db,
	}
}

func (service *SummaryService) GetSummary(id string) (*memory.Summary, error) {
	return service.db.GetSummary(id)
}

func (service *SummaryService) GetSummaries(filter store.Filter) ([]*memory.Summary, error) {
	return service.db.ListSummaries(filter)
}

func (service *SummaryService) DeleteSummary(id string) error {
	return service.db.DeleteSummary(id)
}
