package store

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/DanielStefanK/fitx-utilization-proxy/responses"
	"github.com/jellydator/ttlcache/v3"
)

type Store struct {
	cache           *ttlcache.Cache[uint64, *responses.UtilizationResponse]
	studioListCache []responses.StudioInfo
}

func NewStore() Store {
	cache := ttlcache.New(
		ttlcache.WithTTL[uint64, *responses.UtilizationResponse](15 * time.Minute),
	)

	return Store{
		cache: cache,
	}
}

var httpClient = http.Client{
	Timeout: time.Second * 5, // Timeout after 2 seconds
}

// Get: get a utilization for a studio
func (s *Store) Get(studioId uint64) *responses.UtilizationResponse {

	if s.cache.Get(studioId) == nil {
		magicLineId := s.findMagicLineIdByStudioId(studioId)

		req, err := http.NewRequest(http.MethodGet, fmt.Sprintf("https://mein.fitx.de/nox/public/v1/studios/%d/utilization", magicLineId), nil)

		if err != nil {
			log.Print(err)
		}

		req.Header.Set("x-tenant", "fitx")
		req.Header.Set("x-public-facility-group", os.Getenv("FITX_FACILITY_GROUP"))

		res, err := httpClient.Do(req)

		if err != nil {
			log.Print(err)
			return nil
		}

		body, err := ioutil.ReadAll(res.Body)
		if err != nil {
			log.Print(err)
			return nil
		}

		utilization := responses.UtilizationResponse{}
		err = json.Unmarshal(body, &utilization)
		if err != nil {
			log.Print(string(body))
			log.Print(err)
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
	req, err := http.NewRequest(http.MethodGet, "https://mein.fitx.de/sponsorship/v1/public/studios/forwhitelabelportal", nil)

	if err != nil {
		log.Print(err)
		return nil
	}

	req.Header.Set("x-tenant", "fitx")
	req.Header.Set("x-public-facility-group", os.Getenv("FITX_FACILITY_GROUP"))

	res, err := httpClient.Do(req)

	if err != nil {
		log.Print(err)
		return nil
	}

	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		log.Print(err)
		return nil
	}

	if res.StatusCode != 200 {
		log.Print(string(body))
		return nil
	}

	studioResponse := responses.StudioResponse{}
	err = json.Unmarshal(body, &studioResponse)
	if err != nil {
		log.Print(err)
		return nil
	}

	s.studioListCache = studioResponse.Content

	return &studioResponse
}

func (s *Store) GetStudios() *responses.StudioResponse {
	if s.studioListCache == nil || len(s.studioListCache) == 0 {
		s.UpdateStudios()
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

	return 0
}

func (s *Store) findUuidByStudioId(studioId uint64) string {
	for _, current := range s.studioListCache {
		if current.ID == studioId {
			return current.UUID
		}
	}

	return ""
}

func (s *Store) findNameByStudioId(studioId uint64) string {
	for _, current := range s.studioListCache {
		if current.ID == studioId {
			return current.Name
		}
	}

	return ""
}

func (s *Store) getCurrentWorkload(dataPoints []responses.DataPoint) uint8 {
	for _, current := range dataPoints {
		if current.IsCurrent {
			return current.Percentage
		}
	}

	return 0
}
