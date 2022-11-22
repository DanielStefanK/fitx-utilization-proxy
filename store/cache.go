package store

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"time"

	"github.com/DanielStefanK/fitx-utilization-proxy/responses"
	"github.com/jellydator/ttlcache/v3"
	"go.uber.org/zap"
)

type Store struct {
	cache           *ttlcache.Cache[uint64, *responses.UtilizationResponse]
	studioListCache []responses.StudioInfo
	logger          *zap.Logger
}

func NewStore() Store {
	logger, _ := zap.NewProduction()
	cache := ttlcache.New(
		ttlcache.WithTTL[uint64, *responses.UtilizationResponse](15 * time.Minute),
	)

	store := Store{
		cache:  cache,
		logger: logger,
	}

	store.UpdateStudios()

	return store
}

var httpClient = http.Client{
	Timeout: time.Second * 5, // Timeout after 2 seconds
}

// Get: get a utilization for a studio
func (s *Store) Get(studioId uint64) *responses.UtilizationResponse {

	if s.cache.Get(studioId) == nil {

		s.logger.Info("utilization not found or expired. retrieving.. ", zap.Uint64("studioId", studioId))
		magicLineId := s.findMagicLineIdByStudioId(studioId)

		req, err := http.NewRequest(http.MethodGet, fmt.Sprintf("https://mein.fitx.de/nox/public/v1/studios/%d/utilization", magicLineId), nil)

		if err != nil {
			s.logger.Error(err.Error())
			return nil
		}

		req.Header.Set("x-tenant", "fitx")
		req.Header.Set("x-public-facility-group", os.Getenv("FITX_FACILITY_GROUP"))

		res, err := httpClient.Do(req)

		if err != nil {
			s.logger.Error(err.Error())
			return nil
		}

		body, err := ioutil.ReadAll(res.Body)
		if err != nil {
			s.logger.Error(err.Error())
			return nil
		}

		if res.StatusCode > 399 {
			s.logger.Error("fitx api served an error code", zap.Int("code", res.StatusCode), zap.String("body", string(body)))
			return nil
		}

		utilization := responses.UtilizationResponse{}
		err = json.Unmarshal(body, &utilization)
		if err != nil {
			s.logger.Error(err.Error())
			return nil
		}

		utilization.UUID = s.findUuidByStudioId(studioId)
		utilization.Workload = s.getCurrentWorkload(utilization.Items)
		utilization.Name = s.findNameByStudioId(studioId)

		s.cache.Set(studioId, &utilization, ttlcache.DefaultTTL)

		return &utilization
	} else {
		return s.cache.Get(studioId).Value()
	}
}

func (s *Store) UpdateStudios() *responses.StudioResponse {
	s.logger.Info("Updating studio infos")
	req, err := http.NewRequest(http.MethodGet, "https://mein.fitx.de/sponsorship/v1/public/studios/forwhitelabelportal", nil)

	if err != nil {
		s.logger.Error(err.Error())
		return nil
	}

	req.Header.Set("x-tenant", "fitx")
	req.Header.Set("x-public-facility-group", os.Getenv("FITX_FACILITY_GROUP"))

	res, err := httpClient.Do(req)

	if err != nil {
		s.logger.Error(err.Error())
		return nil
	}

	body, err := ioutil.ReadAll(res.Body)

	if err != nil {
		s.logger.Error(err.Error())
		return nil
	}

	if res.StatusCode > 399 {
		s.logger.Error("fitx api served an error code", zap.Int("code", res.StatusCode), zap.String("body", string(body)))
		return nil
	}

	studioResponse := responses.StudioResponse{}
	err = json.Unmarshal(body, &studioResponse)
	if err != nil {
		s.logger.Error(err.Error())
		return nil
	}

	s.studioListCache = studioResponse.Content

	return &studioResponse
}

func (s *Store) GetStudios() *responses.StudioResponse {
	if s.studioListCache == nil || len(s.studioListCache) == 0 {
		studios := s.UpdateStudios()

		if studios == nil {
			return nil
		}
	}

	return &responses.StudioResponse{
		Content: s.studioListCache,
	}
}

func (s *Store) findMagicLineIdByStudioId(studioId uint64) uint64 {
	for _, current := range s.studioListCache {
		if current.ID == studioId {
			return current.MagiclineId
		}
	}

	s.logger.Warn("Could not find magiclinId by studioId",
		zap.Uint64("studioId", studioId),
	)

	return 0
}

func (s *Store) findUuidByStudioId(studioId uint64) string {
	for _, current := range s.studioListCache {
		if current.ID == studioId {
			return current.UUID
		}
	}

	s.logger.Warn("Could not find uuid by studioId",
		zap.Uint64("studioId", studioId),
	)

	return ""
}

func (s *Store) findNameByStudioId(studioId uint64) string {
	for _, current := range s.studioListCache {
		if current.ID == studioId {
			return current.Name
		}
	}

	s.logger.Warn("Could not find studio name by studioId",
		zap.Uint64("studioId", studioId),
	)

	return ""
}

func (s *Store) getCurrentWorkload(dataPoints []responses.DataPoint) uint8 {
	for _, current := range dataPoints {
		if current.IsCurrent {
			return current.Percentage
		}
	}

	s.logger.Error("could not find current workload")
	return 0
}

func (s *Store) StudioExists(studioId uint64) bool {
	for _, current := range s.studioListCache {
		if current.ID == studioId {
			return true
		}
	}

	return false
}
