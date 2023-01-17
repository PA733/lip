// Package metadata contains the metadata of a tooth.
package toothmetadata

import (
	"bytes"
	"encoding/json"
	"errors"
	"regexp"
	"strings"

	"github.com/liteldev/lip/tooth"
	versionutils "github.com/liteldev/lip/utils/version"
	"github.com/liteldev/lip/utils/version/versionmatch"
)

// InfoStruct is the struct that contains the information of a tooth.
type InfoStruct struct {
	Name        string
	Description string
	Author      string
	License     string
	Homepage    string
}

// PlacementStruct is the struct that contains the source and destination of a placement.
type PlacementStruct struct {
	Source      string
	Destination string
}

// Metadata is the struct that contains all the metadata of a tooth.
type Metadata struct {
	ToothPath    string
	Version      versionutils.Version
	Dependencies map[string]([][]versionmatch.VersionMatch)
	Information  InfoStruct
	Placement    []PlacementStruct
}

// NewFromJSON decodes a JSON byte array into a Metadata struct.
func NewFromJSON(jsonData []byte) (Metadata, error) {
	// Read to a map.
	var metadataMap map[string]interface{}
	err := json.Unmarshal(jsonData, &metadataMap)
	if err != nil {
		return Metadata{}, errors.New("failed to decode JSON into metadata: " + err.Error())
	}

	// Parse to metadata.
	var metadata Metadata

	// Tooth path should be lower case.
	metadata.ToothPath = strings.ToLower(metadataMap["tooth"].(string))
	if !tooth.IsValidToothPath(metadata.ToothPath) {
		return Metadata{}, errors.New("failed to decode JSON into metadata: invalid tooth path: " + metadata.ToothPath)
	}

	version, err := versionutils.NewFromString(metadataMap["version"].(string))

	if err != nil {
		return Metadata{}, errors.New("failed to decode JSON into metadata: " + err.Error())
	}
	metadata.Version = version

	metadata.Dependencies = make(map[string]([][]versionmatch.VersionMatch))
	for toothPath, versionMatchOuterList := range metadataMap["dependencies"].(map[string]interface{}) {
		metadata.Dependencies[toothPath] = make([][]versionmatch.VersionMatch, len(versionMatchOuterList.([]interface{})))
		for i, versionMatchInnerList := range versionMatchOuterList.([]interface{}) {
			metadata.Dependencies[toothPath][i] = make([]versionmatch.VersionMatch, len(versionMatchInnerList.([]interface{})))
			for j, versionMatch := range versionMatchInnerList.([]interface{}) {
				versionMatch, err := versionmatch.NewFromString(versionMatch.(string))
				if err != nil {
					return Metadata{}, errors.New("failed to decode JSON into metadata: " + err.Error())
				}

				metadata.Dependencies[toothPath][i][j] = versionMatch
			}
		}
	}

	metadata.Information.Name = metadataMap["information"].(map[string]interface{})["name"].(string)
	metadata.Information.Description = metadataMap["information"].(map[string]interface{})["description"].(string)
	metadata.Information.Author = metadataMap["information"].(map[string]interface{})["author"].(string)
	metadata.Information.License = metadataMap["information"].(map[string]interface{})["license"].(string)
	metadata.Information.Homepage = metadataMap["information"].(map[string]interface{})["homepage"].(string)

	metadata.Placement = make([]PlacementStruct, len(metadataMap["placement"].([]interface{})))
	for i, placement := range metadataMap["placement"].([]interface{}) {
		source := placement.(map[string]interface{})["source"].(string)
		destination := placement.(map[string]interface{})["destination"].(string)

		// Source and destination should starts with a letter or a digit and should only contains
		reg := regexp.MustCompile(`^[a-zA-Z0-9]\S*$`)
		// The matched string should be the same as the original string.
		if reg.FindString(source) != source {
			return Metadata{}, errors.New("failed to decode JSON into metadata: invalid source: " + source)
		}
		if reg.FindString(destination) != destination {
			return Metadata{}, errors.New("failed to decode JSON into metadata: invalid destination: " + destination)
		}

		metadata.Placement[i].Source = source
		metadata.Placement[i].Destination = destination
	}

	return metadata, nil
}

// JSON encodes a Metadata struct into a JSON byte array.
func (metadata Metadata) JSON() ([]byte, error) {
	metadataMap := make(map[string]interface{})

	metadataMap["tooth"] = metadata.ToothPath

	metadataMap["version"] = metadata.Version.String()

	metadataMap["dependencies"] = make(map[string]interface{})
	for toothPath, versionMatchOuterList := range metadata.Dependencies {
		metadataMap["dependencies"].(map[string]interface{})[toothPath] =
			make([]interface{}, len(versionMatchOuterList))
		for i, versionMatchInnerList := range versionMatchOuterList {
			metadataMap["dependencies"].(map[string]interface{})[toothPath].([]interface{})[i] =
				make([]interface{}, len(versionMatchInnerList))
			for j, versionMatch := range versionMatchInnerList {
				metadataMap["dependencies"].(map[string]interface{})[toothPath].([]interface{})[i].([]interface{})[j] = versionMatch.String()
			}
		}
	}

	metadataMap["information"] = make(map[string]interface{})
	metadataMap["information"].(map[string]interface{})["name"] = metadata.Information.Name
	metadataMap["information"].(map[string]interface{})["description"] = metadata.Information.Description
	metadataMap["information"].(map[string]interface{})["author"] = metadata.Information.Author
	metadataMap["information"].(map[string]interface{})["license"] = metadata.Information.License
	metadataMap["information"].(map[string]interface{})["homepage"] = metadata.Information.Homepage

	metadataMap["placement"] = make([]interface{}, len(metadata.Placement))
	for i, placement := range metadata.Placement {
		metadataMap["placement"].([]interface{})[i] = make(map[string]interface{})
		metadataMap["placement"].([]interface{})[i].(map[string]interface{})["source"] = placement.Source
		metadataMap["placement"].([]interface{})[i].(map[string]interface{})["destination"] = placement.Destination
	}

	// Encode metadataMap into JSON
	buf := bytes.NewBuffer([]byte{})
	encoder := json.NewEncoder(buf)

	encoder.SetIndent("", "  ")

	// Prevent HTML escaping. Otherwise, "<", ">", "&", U+2028, and U+2029
	// characters are escaped to "\u003c", "\u003e", "\u0026", "\u2028", and "\u2029".
	encoder.SetEscapeHTML(false)

	err := encoder.Encode(metadataMap)
	if err != nil {
		return nil, errors.New("failed to encode metadata into JSON: " + err.Error())
	}

	return buf.Bytes(), nil
}
