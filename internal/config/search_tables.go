package config

import (
	"fmt"
	"regexp"
)

var establishmentsTablePattern = regexp.MustCompile(
	`^cnpj_establishments(_[a-z0-9_]{1,48})?$`,
)

var partnersTablePattern = regexp.MustCompile(
	`^cnpj_partners(_[a-z0-9_]{1,48})?$`,
)

func validateManticoreEntityTables(configuration Config) error {
	if !establishmentsTablePattern.MatchString(
		configuration.ManticoreEstablishmentsTable,
	) {
		return fmt.Errorf(
			"%s must be cnpj_establishments or cnpj_establishments_ followed by 1 to 48 lowercase letters, digits or underscores",
			manticoreEstablishmentsTableEnvironmentVariable,
		)
	}
	if !partnersTablePattern.MatchString(
		configuration.ManticorePartnersTable,
	) {
		return fmt.Errorf(
			"%s must be cnpj_partners or cnpj_partners_ followed by 1 to 48 lowercase letters, digits or underscores",
			manticorePartnersTableEnvironmentVariable,
		)
	}
	return nil
}

func (configuration Config) IsActiveSearchTable(table string) bool {
	if table == "" {
		return false
	}
	return table == configuration.ManticoreTable ||
		table == configuration.ManticoreEstablishmentsTable ||
		table == configuration.ManticorePartnersTable
}
