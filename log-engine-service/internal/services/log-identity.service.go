package services

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
)

type LogIdentityService struct{}

func NewLogIdentityService() *LogIdentityService {
	return &LogIdentityService{}
}

func (s *LogIdentityService) BuildLogIDFromRawData(projectID uint64, applicationID uint64, sourceIP string, rawData []byte) string {
	rawHash := sha256.Sum256(rawData)
	return s.BuildLogIDFromRawHash(projectID, applicationID, sourceIP, hex.EncodeToString(rawHash[:]))
}

func (s *LogIdentityService) BuildLogIDFromRawHash(projectID uint64, applicationID uint64, sourceIP string, rawDataSHA256 string) string {
	canonical := fmt.Sprintf(
		"v1|%d|%d|%s|%s",
		projectID,
		applicationID,
		sourceIP,
		rawDataSHA256,
	)

	id := sha256.Sum256([]byte(canonical))
	return hex.EncodeToString(id[:])
}
