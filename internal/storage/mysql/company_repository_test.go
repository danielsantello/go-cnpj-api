package mysqlstorage

import "testing"

func TestNewCompanyRepositoryAcceptsVersion1(t *testing.T) {
	repository, err := NewCompanyRepository(nil, 1)
	if err != nil {
		t.Fatalf("create company repository: %v", err)
	}

	if repository == nil {
		t.Fatal("expected company repository, got nil")
	}
}

func TestNewCompanyRepositoryRejectsUnsupportedVersion(t *testing.T) {
	repository, err := NewCompanyRepository(nil, 999)

	if err == nil {
		t.Fatal("expected unsupported version error, got nil")
	}

	if repository != nil {
		t.Fatalf("expected nil repository, got %#v", repository)
	}
}
