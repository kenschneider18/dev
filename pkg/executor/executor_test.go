package executor

import "testing"

func TestNormalizeClonePath(t *testing.T) {
	tests := []struct {
		name          string
		input         string
		wantRepoPath  string
		wantClonePath string
		wantErr       bool
	}{
		{
			name:          "plain path",
			input:         "github.com/kenschneider18/dev",
			wantRepoPath:  "github.com/kenschneider18/dev",
			wantClonePath: "https://github.com/kenschneider18/dev",
		},
		{
			name:          "plain path with .git",
			input:         "github.com/kenschneider18/dev.git",
			wantRepoPath:  "github.com/kenschneider18/dev",
			wantClonePath: "https://github.com/kenschneider18/dev",
		},
		{
			name:          "https URL",
			input:         "https://github.com/kenschneider18/dev",
			wantRepoPath:  "github.com/kenschneider18/dev",
			wantClonePath: "https://github.com/kenschneider18/dev",
		},
		{
			name:          "https URL with .git",
			input:         "https://github.com/kenschneider18/dev.git",
			wantRepoPath:  "github.com/kenschneider18/dev",
			wantClonePath: "https://github.com/kenschneider18/dev.git",
		},
		{
			name:          "http URL",
			input:         "http://gitlab.com/group/subgroup/project.git",
			wantRepoPath:  "gitlab.com/group/subgroup/project",
			wantClonePath: "http://gitlab.com/group/subgroup/project.git",
		},
		{
			name:          "bitbucket plain path",
			input:         "bitbucket.org/team/project",
			wantRepoPath:  "bitbucket.org/team/project",
			wantClonePath: "https://bitbucket.org/team/project",
		},
		{
			name:          "gitlab https URL",
			input:         "https://gitlab.com/group/project.git",
			wantRepoPath:  "gitlab.com/group/project",
			wantClonePath: "https://gitlab.com/group/project.git",
		},
		{
			name:          "https URL with token",
			input:         "https://token@gitlab.com/group/project.git",
			wantRepoPath:  "gitlab.com/group/project",
			wantClonePath: "https://token@gitlab.com/group/project.git",
		},
		{
			name:          "ssh URL",
			input:         "git@github.com:kenschneider18/dev.git",
			wantRepoPath:  "github.com/kenschneider18/dev",
			wantClonePath: "git@github.com:kenschneider18/dev.git",
		},
		{
			name:          "ssh URL slash separator",
			input:         "git@github.com/kenschneider18/dev",
			wantRepoPath:  "github.com/kenschneider18/dev",
			wantClonePath: "git@github.com/kenschneider18/dev",
		},
		{
			name:    "invalid https URL missing path",
			input:   "https://github.com",
			wantErr: true,
		},
		{
			name:    "invalid http URL missing path",
			input:   "http://bitbucket.org",
			wantErr: true,
		},
		{
			name:    "invalid ssh URL missing path",
			input:   "git@github.com",
			wantErr: true,
		},
		{
			name:    "invalid plain path",
			input:   "github.com/kenschneider18",
			wantErr: true,
		},
		{
			name:    "empty string",
			input:   "",
			wantErr: true,
		},
		{
			name:    "whitespace only",
			input:   "   ",
			wantErr: true,
		},
		{
			name:    "git@ with no content",
			input:   "git@",
			wantErr: true,
		},
	}

	for _, test := range tests {
		test := test
		t.Run(test.name, func(t *testing.T) {
			gotRepoPath, gotClonePath, err := normalizeClonePath(test.input)
			if (err != nil) != test.wantErr {
				t.Fatalf("normalizeClonePath() error = %v, wantErr %v", err, test.wantErr)
			}

			if test.wantErr {
				return
			}

			if gotRepoPath != test.wantRepoPath {
				t.Fatalf("normalizeClonePath() repoPath = %q, want %q", gotRepoPath, test.wantRepoPath)
			}

			if gotClonePath != test.wantClonePath {
				t.Fatalf("normalizeClonePath() clonePath = %q, want %q", gotClonePath, test.wantClonePath)
			}
		})
	}
}
