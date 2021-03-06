package remoteworkitem

import (
	"testing"

	"golang.org/x/net/context"

	"github.com/almighty/almighty-core/app"
	"github.com/almighty/almighty-core/application"
	"github.com/almighty/almighty-core/criteria"
	"github.com/almighty/almighty-core/gormsupport/cleaner"
	"github.com/almighty/almighty-core/gormtestsupport"
	"github.com/almighty/almighty-core/resource"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

type TestTrackerRepository struct {
	gormtestsupport.DBTestSuite

	repo application.TrackerRepository

	clean func()
}

func TestRunTrackerRepository(t *testing.T) {
	suite.Run(t, &TestTrackerRepository{DBTestSuite: gormtestsupport.NewDBTestSuite("../config.yaml")})
}

func (test *TestTrackerRepository) SetupTest() {
	test.repo = NewTrackerRepository(test.DB)
	test.clean = cleaner.DeleteCreatedEntities(test.DB)
}

func (test *TestTrackerRepository) TearDownTest() {
	test.clean()
}

func (test *TestTrackerRepository) TestTrackerCreate() {
	t := test.T()
	resource.Require(t, resource.Database)

	tracker, err := test.repo.Create(context.Background(), "gugus", "dada")
	assert.IsType(t, BadParameterError{}, err)
	assert.Nil(t, tracker)

	tracker, err = test.repo.Create(context.Background(), "http://api.github.com", ProviderGithub)
	assert.Nil(t, err)
	assert.NotNil(t, tracker)
	assert.Equal(t, "http://api.github.com", tracker.URL)
	assert.Equal(t, ProviderGithub, tracker.Type)

	tracker2, err := test.repo.Load(context.Background(), tracker.ID)
	assert.Nil(t, err)
	assert.NotNil(t, tracker2)
}

func (test *TestTrackerRepository) TestTrackerSave() {
	t := test.T()
	resource.Require(t, resource.Database)

	tracker, err := test.repo.Save(context.Background(), app.Tracker{})
	assert.IsType(t, NotFoundError{}, err)
	assert.Nil(t, tracker)

	tracker, _ = test.repo.Create(context.Background(), "http://api.github.com", ProviderGithub)
	tracker.Type = "blabla"
	tracker2, err := test.repo.Save(context.Background(), *tracker)
	assert.IsType(t, BadParameterError{}, err)
	assert.Nil(t, tracker2)

	tracker.Type = ProviderJira
	tracker.URL = "blabla"
	tracker, err = test.repo.Save(context.Background(), *tracker)
	assert.Equal(t, ProviderJira, tracker.Type)
	assert.Equal(t, "blabla", tracker.URL)

	tracker.ID = "10000"
	tracker2, err = test.repo.Save(context.Background(), *tracker)
	assert.IsType(t, NotFoundError{}, err)
	assert.Nil(t, tracker2)

	tracker.ID = "asdf"
	tracker2, err = test.repo.Save(context.Background(), *tracker)
	assert.IsType(t, NotFoundError{}, err)
	assert.Nil(t, tracker2)
}

func (test *TestTrackerRepository) TestTrackerDelete() {
	t := test.T()
	resource.Require(t, resource.Database)

	err := test.repo.Delete(context.Background(), "asdf")
	assert.IsType(t, NotFoundError{}, err)

	// guard against other test leaving stuff behind
	err = test.repo.Delete(context.Background(), "10000")

	err = test.repo.Delete(context.Background(), "10000")
	assert.IsType(t, NotFoundError{}, err)

	tracker, _ := test.repo.Create(context.Background(), "http://api.github.com", ProviderGithub)
	err = test.repo.Delete(context.Background(), tracker.ID)
	assert.Nil(t, err)

	tracker, err = test.repo.Load(context.Background(), tracker.ID)
	assert.IsType(t, NotFoundError{}, err)
	assert.Nil(t, tracker)

	tracker, err = test.repo.Load(context.Background(), "xyz")
	assert.IsType(t, NotFoundError{}, err)
	assert.Nil(t, tracker)
}

func (test *TestTrackerRepository) TestTrackerList() {
	t := test.T()
	resource.Require(t, resource.Database)

	trackers, _ := test.repo.List(context.Background(), criteria.Literal(true), nil, nil)

	test.repo.Create(context.Background(), "http://api.github.com", ProviderGithub)
	test.repo.Create(context.Background(), "http://issues.jboss.com", ProviderJira)
	test.repo.Create(context.Background(), "http://issues.jboss.com", ProviderJira)
	test.repo.Create(context.Background(), "http://api.github.com", ProviderGithub)

	trackers2, _ := test.repo.List(context.Background(), criteria.Literal(true), nil, nil)

	assert.Equal(t, len(trackers)+4, len(trackers2))
	start, len := 1, 1

	trackers3, _ := test.repo.List(context.Background(), criteria.Literal(true), &start, &len)
	assert.Equal(t, trackers2[1], trackers3[0])
}
