package webhook

import (
	"encoding/json"
	"net/http"

	"github.com/KrushnaVardhanReddy/substrate/api/internal/db"
	"github.com/KrushnaVardhanReddy/substrate/api/internal/github"
)

type ReactionPayload struct {
	Action  string `json:"action"`
	Comment struct {
		ID   int64  `json:"id"`
		Body string `json:"body"`
	} `json:"comment"`
	Reaction struct {
		Content string `json:"content"`
	} `json:"reaction"`
	Sender struct {
		Login string `json:"login"`
	} `json:"sender"`
	Repository struct {
		Name  string `json:"name"`
		Owner struct {
			Login string `json:"login"`
		} `json:"owner"`
	} `json:"repository"`
	Issue struct {
		Number int `json:"number"`
	} `json:"issue"`
}

func ReactionHandler(store db.Store, ghClient github.Client) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req ReactionPayload
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, `{"error": "invalid json"}`, http.StatusBadRequest)
			return
		}

		if req.Action == "created" && req.Reaction.Content == "+1" {
			// This is a 👍 reaction

			// Check if sender is mentioned in the comment
			mentions := github.ParseNegotiationComment(req.Comment.Body)
			senderMention := "@" + req.Sender.Login
			isMentioned := false
			for _, m := range mentions {
				if m == senderMention {
					isMentioned = true
					break
				}
			}

			if isMentioned {
				// Get all reactions on this comment to see if all mentioned users have approved
				approvals, err := ghClient.GetIssueCommentReactions(r.Context(), req.Repository.Owner.Login, req.Repository.Name, req.Issue.Number, req.Comment.ID)

				allApproved := true
				if err == nil {
					approvalMap := make(map[string]bool)
					for _, u := range approvals {
						approvalMap[u] = true
					}
					for _, m := range mentions {
						if !approvalMap[m] {
							allApproved = false
							break
						}
					}
				} else {
					allApproved = false
				}

				if allApproved {
					commitSHA, localErr := ghClient.GetPullRequestHeadSHA(r.Context(), req.Repository.Owner.Login, req.Repository.Name, req.Issue.Number)
					if localErr == nil {
						_ = ghClient.CreateCheckRun(
							r.Context(),
							req.Repository.Owner.Login,
							req.Repository.Name,
							commitSHA,
							"Substrate Contract Negotiation",
							"Consumer Approval Received",
							"All affected consumers have approved the breaking change.",
							"success",
						)
					}
				}
			}

		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusAccepted)
		json.NewEncoder(w).Encode(map[string]string{"status": "processed"})
	}
}
