package utils

import (
	"encoding/json"
	"os"

	"github.com/rs/zerolog/log"
)

func CheckErr(err error) {
	if err != nil {
		log.Error().Err(err).Send()
	}
}

func ReadJsonFile(filePath string, value interface{}) error {
	if content, err := os.ReadFile(filePath); err != nil {
		return err
	} else {
		if err = json.Unmarshal(content, value); err != nil {
			return err
		}
	}

	return nil
}

func SaveJsonFile(filePath string, value interface{}) error {
	if data, err := json.Marshal(value); err != nil {
		return err
	} else {
		if err = os.WriteFile(filePath, data, 0644); err != nil {
			return err
		}
	}

	return nil
}

func UniqueStringSlice(s []string) []string {
	uniqueSlice := make([]string, 0)
	uniqueMap := map[string]bool{}

	for _, val := range s {
		if uniqueMap[val] {
			continue
		}
		uniqueMap[val] = true
		uniqueSlice = append(uniqueSlice, val)
	}

	return uniqueSlice
}
