package service_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/krzachariassen/unm-platform/internal/domain/entity"
	"github.com/krzachariassen/unm-platform/internal/domain/service"
)

func makeModel(actors, needs, caps, svcs, teams []string) *entity.UNMModel {
	m := entity.NewUNMModel("", "")
	for _, name := range actors {
		_ = m.AddActor(&entity.Actor{Name: name})
	}
	for _, name := range needs {
		_ = m.AddNeed(&entity.Need{Name: name})
	}
	for _, name := range caps {
		_ = m.AddCapability(&entity.Capability{Name: name})
	}
	for _, name := range svcs {
		_ = m.AddService(&entity.Service{Name: name})
	}
	for _, name := range teams {
		_ = m.AddTeam(&entity.Team{Name: name})
	}
	return m
}

func TestDiff_NoChanges(t *testing.T) {
	m := makeModel([]string{"A"}, []string{"N"}, []string{"C"}, []string{"S"}, []string{"T"})
	diff := service.Diff(m, m)
	assert.Empty(t, diff.Added.Actors)
	assert.Empty(t, diff.Removed.Actors)
	assert.Empty(t, diff.Changed.Actors)
}

func TestDiff_Added(t *testing.T) {
	from := makeModel([]string{"A"}, nil, nil, nil, nil)
	to := makeModel([]string{"A", "B"}, nil, []string{"NewCap"}, nil, []string{"Team X"})
	diff := service.Diff(from, to)
	assert.Equal(t, []string{"B"}, diff.Added.Actors)
	assert.Equal(t, []string{"NewCap"}, diff.Added.Capabilities)
	assert.Equal(t, []string{"Team X"}, diff.Added.Teams)
	assert.Empty(t, diff.Removed.Actors)
}

func TestDiff_Removed(t *testing.T) {
	from := makeModel([]string{"A", "B"}, nil, nil, []string{"Svc1"}, nil)
	to := makeModel([]string{"A"}, nil, nil, nil, nil)
	diff := service.Diff(from, to)
	assert.Equal(t, []string{"B"}, diff.Removed.Actors)
	assert.Equal(t, []string{"Svc1"}, diff.Removed.Services)
	assert.Empty(t, diff.Added.Actors)
}

func TestDiff_NilModels(t *testing.T) {
	diff := service.Diff(nil, nil)
	assert.NotNil(t, diff)
	assert.Empty(t, diff.Added.Actors)
}
