package git

import (
	"github.com/go-git/go-git/v5"
)

func GetCurrentSHA(path string) (string, error) {
	repo, err := git.PlainOpen(path)
	if err != nil {
		return "local", nil // Fallback for non-git
	}

	head, err := repo.Head()
	if err != nil {
		return "local", nil
	}

	return head.Hash().String(), nil
}

func GetParentSHA(path string) (string, error) {
	repo, err := git.PlainOpen(path)
	if err != nil {
		return "local", nil
	}

	head, err := repo.Head()
	if err != nil {
		return "local", nil
	}

	commit, err := repo.CommitObject(head.Hash())
	if err != nil {
		return "local", nil
	}

	if commit.NumParents() == 0 {
		return head.Hash().String(), nil
	}

	parent, err := commit.Parent(0)
	if err != nil {
		return head.Hash().String(), nil
	}

	return parent.Hash.String(), nil
}
