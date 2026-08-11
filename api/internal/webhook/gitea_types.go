// Copyright (c) 2026 Krushna Vardhan Reddy. All rights reserved.
// PROPRIETARY AND CONFIDENTIAL — Substrate Cloud Platform.
// Unauthorized copying, modification, or distribution is strictly prohibited.

package webhook

// GiteaPushWebhook represents the payload sent by Gitea/Forgejo on a push event
type GiteaPushWebhook struct {
	Secret     string          `json:"secret"`
	Ref        string          `json:"ref"`
	Before     string          `json:"before"`
	After      string          `json:"after"`
	Repository GiteaRepository `json:"repository"`
	Commits    []GiteaCommit   `json:"commits"`
}

type GiteaRepository struct {
	ID       int64      `json:"id"`
	Name     string     `json:"name"`
	FullName string     `json:"full_name"`
	Owner    GiteaOwner `json:"owner"`
}

type GiteaOwner struct {
	ID    int64  `json:"id"`
	Login string `json:"login"`
}

type GiteaCommit struct {
	ID       string   `json:"id"`
	Added    []string `json:"added"`
	Removed  []string `json:"removed"`
	Modified []string `json:"modified"`
}
