package proofchain

import (
	"context"
	"fmt"
	"net/url"
	"time"
)

// Quest represents a quest definition
type Quest struct {
	ID                    string      `json:"id"`
	Name                  string      `json:"name"`
	Slug                  string      `json:"slug"`
	Description           *string     `json:"description,omitempty"`
	ShortDescription      *string     `json:"short_description,omitempty"`
	IconURL               *string     `json:"icon_url,omitempty"`
	BannerURL             *string     `json:"banner_url,omitempty"`
	Category              *string     `json:"category,omitempty"`
	Difficulty            *string     `json:"difficulty,omitempty"`
	EstimatedTime         *string     `json:"estimated_time,omitempty"`
	IsOrdered             bool        `json:"is_ordered"`
	IsRepeatable          bool        `json:"is_repeatable"`
	RepeatCooldownHours   *int        `json:"repeat_cooldown_hours,omitempty"`
	MaxCompletionsPerUser *int        `json:"max_completions_per_user,omitempty"`
	StartsAt              *time.Time  `json:"starts_at,omitempty"`
	EndsAt                *time.Time  `json:"ends_at,omitempty"`
	TimeLimitHours        *int        `json:"time_limit_hours,omitempty"`
	PrerequisiteQuestIDs  []string    `json:"prerequisite_quest_ids"`
	MaxParticipants       *int        `json:"max_participants,omitempty"`
	MaxCompletions        *int        `json:"max_completions,omitempty"`
	RewardDefinitionID    *string     `json:"reward_definition_id,omitempty"`
	RewardPoints          *int        `json:"reward_points,omitempty"`
	IsPublic              bool        `json:"is_public"`
	IsFeatured            bool        `json:"is_featured"`
	Tags                  []string    `json:"tags"`
	Status                string      `json:"status"`
	Steps                 []QuestStep `json:"steps"`
	TotalParticipants     int         `json:"total_participants"`
	TotalCompletions      int         `json:"total_completions"`
	CreatedAt             time.Time   `json:"created_at"`
	UpdatedAt             time.Time   `json:"updated_at"`
}

// QuestStep represents a step in a quest
type QuestStep struct {
	ID                 string                 `json:"id"`
	QuestID            string                 `json:"quest_id"`
	Name               string                 `json:"name"`
	Description        *string                `json:"description,omitempty"`
	Order              int                    `json:"order"`
	StepType           string                 `json:"step_type"`
	EventType          *string                `json:"event_type,omitempty"`
	EventTypes         []string               `json:"event_types,omitempty"`
	Criteria           map[string]interface{} `json:"criteria,omitempty"`
	RequiredDataFields []string               `json:"required_data_fields,omitempty"`
	StepPoints         *int                   `json:"step_points,omitempty"`
	IconURL            *string                `json:"icon_url,omitempty"`
	IsOptional         bool                   `json:"is_optional"`
}

// UserQuestProgress represents a user's progress on a quest
type UserQuestProgress struct {
	ID               string         `json:"id"`
	UserID           string         `json:"user_id"`
	QuestID          string         `json:"quest_id"`
	QuestName        string         `json:"quest_name"`
	Status           string         `json:"status"`
	StartedAt        *time.Time     `json:"started_at,omitempty"`
	CompletedAt      *time.Time     `json:"completed_at,omitempty"`
	ExpiresAt        *time.Time     `json:"expires_at,omitempty"`
	CompletionCount  int            `json:"completion_count"`
	CurrentStepOrder int            `json:"current_step_order"`
	StepProgress     []StepProgress `json:"step_progress"`
	PointsEarned     int            `json:"points_earned"`
	RewardEarned     bool           `json:"reward_earned"`
}

// StepProgress represents progress on a single step
type StepProgress struct {
	StepID      string     `json:"step_id"`
	StepName    string     `json:"step_name"`
	Order       int        `json:"order"`
	Status      string     `json:"status"`
	CompletedAt *time.Time `json:"completed_at,omitempty"`
	EventID     *string    `json:"event_id,omitempty"`
}

// QuestWithProgress includes user progress
type QuestWithProgress struct {
	Quest
	UserProgress *UserQuestProgress `json:"user_progress,omitempty"`
}

// Request types
type CreateQuestRequest struct {
	Name                  string                   `json:"name"`
	Slug                  string                   `json:"slug,omitempty"`
	Description           *string                  `json:"description,omitempty"`
	ShortDescription      *string                  `json:"short_description,omitempty"`
	IconURL               *string                  `json:"icon_url,omitempty"`
	BannerURL             *string                  `json:"banner_url,omitempty"`
	Category              *string                  `json:"category,omitempty"`
	Difficulty            *string                  `json:"difficulty,omitempty"`
	EstimatedTime         *string                  `json:"estimated_time,omitempty"`
	IsOrdered             bool                     `json:"is_ordered,omitempty"`
	IsRepeatable          bool                     `json:"is_repeatable,omitempty"`
	RepeatCooldownHours   *int                     `json:"repeat_cooldown_hours,omitempty"`
	MaxCompletionsPerUser *int                     `json:"max_completions_per_user,omitempty"`
	StartsAt              *time.Time               `json:"starts_at,omitempty"`
	EndsAt                *time.Time               `json:"ends_at,omitempty"`
	TimeLimitHours        *int                     `json:"time_limit_hours,omitempty"`
	PrerequisiteQuestIDs  []string                 `json:"prerequisite_quest_ids,omitempty"`
	MaxParticipants       *int                     `json:"max_participants,omitempty"`
	MaxCompletions        *int                     `json:"max_completions,omitempty"`
	RewardDefinitionID    *string                  `json:"reward_definition_id,omitempty"`
	RewardPoints          *int                     `json:"reward_points,omitempty"`
	IsPublic              bool                     `json:"is_public,omitempty"`
	IsFeatured            bool                     `json:"is_featured,omitempty"`
	Tags                  []string                 `json:"tags,omitempty"`
	Steps                 []CreateQuestStepRequest `json:"steps"`
}

type CreateQuestStepRequest struct {
	Name               string                 `json:"name"`
	Description        *string                `json:"description,omitempty"`
	Order              *int                   `json:"order,omitempty"`
	StepType           string                 `json:"step_type,omitempty"`
	EventType          *string                `json:"event_type,omitempty"`
	EventTypes         []string               `json:"event_types,omitempty"`
	Criteria           map[string]interface{} `json:"criteria,omitempty"`
	RequiredDataFields []string               `json:"required_data_fields,omitempty"`
	StepPoints         *int                   `json:"step_points,omitempty"`
	IconURL            *string                `json:"icon_url,omitempty"`
	IsOptional         bool                   `json:"is_optional,omitempty"`
}

type ListQuestsOptions struct {
	Status     string
	Category   string
	IsPublic   *bool
	IsFeatured *bool
	Limit      int
	Offset     int
}

// QuestsClient provides quest operations
type QuestsClient struct {
	http *HTTPClient
}

// NewQuestsClient creates a new quests client
func NewQuestsClient(http *HTTPClient) *QuestsClient {
	return &QuestsClient{http: http}
}

// List returns quests
func (q *QuestsClient) List(ctx context.Context, opts *ListQuestsOptions) ([]Quest, error) {
	params := url.Values{}
	if opts != nil {
		if opts.Status != "" {
			params.Set("status", opts.Status)
		}
		if opts.Category != "" {
			params.Set("category", opts.Category)
		}
		if opts.IsPublic != nil {
			params.Set("is_public", fmt.Sprintf("%t", *opts.IsPublic))
		}
		if opts.IsFeatured != nil {
			params.Set("is_featured", fmt.Sprintf("%t", *opts.IsFeatured))
		}
		if opts.Limit > 0 {
			params.Set("limit", fmt.Sprintf("%d", opts.Limit))
		}
		if opts.Offset > 0 {
			params.Set("offset", fmt.Sprintf("%d", opts.Offset))
		}
	}

	var quests []Quest
	err := q.http.Get(ctx, "/quests", params, &quests)
	return quests, err
}

// Get returns a quest by ID
func (q *QuestsClient) Get(ctx context.Context, questID string) (*Quest, error) {
	var quest Quest
	err := q.http.Get(ctx, "/quests/"+questID, nil, &quest)
	if err != nil {
		return nil, err
	}
	return &quest, nil
}

// GetBySlug returns a quest by slug
func (q *QuestsClient) GetBySlug(ctx context.Context, slug string) (*Quest, error) {
	var quest Quest
	err := q.http.Get(ctx, "/quests/slug/"+url.PathEscape(slug), nil, &quest)
	if err != nil {
		return nil, err
	}
	return &quest, nil
}

// Create creates a quest
func (q *QuestsClient) Create(ctx context.Context, req *CreateQuestRequest) (*Quest, error) {
	var quest Quest
	err := q.http.Post(ctx, "/quests", req, &quest)
	if err != nil {
		return nil, err
	}
	return &quest, nil
}

// Update updates a quest
func (q *QuestsClient) Update(ctx context.Context, questID string, req *CreateQuestRequest) (*Quest, error) {
	var quest Quest
	err := q.http.Put(ctx, "/quests/"+questID, req, &quest)
	if err != nil {
		return nil, err
	}
	return &quest, nil
}

// Delete deletes a quest
func (q *QuestsClient) Delete(ctx context.Context, questID string) error {
	return q.http.Delete(ctx, "/quests/"+questID)
}

// Activate activates a quest
func (q *QuestsClient) Activate(ctx context.Context, questID string) (*Quest, error) {
	var quest Quest
	err := q.http.Post(ctx, "/quests/"+questID+"/activate", nil, &quest)
	if err != nil {
		return nil, err
	}
	return &quest, nil
}

// Pause pauses a quest
func (q *QuestsClient) Pause(ctx context.Context, questID string) (*Quest, error) {
	var quest Quest
	err := q.http.Post(ctx, "/quests/"+questID+"/pause", nil, &quest)
	if err != nil {
		return nil, err
	}
	return &quest, nil
}

// Archive archives a quest
func (q *QuestsClient) Archive(ctx context.Context, questID string) (*Quest, error) {
	var quest Quest
	err := q.http.Post(ctx, "/quests/"+questID+"/archive", nil, &quest)
	if err != nil {
		return nil, err
	}
	return &quest, nil
}

// GetWithProgress returns quest with user progress
func (q *QuestsClient) GetWithProgress(ctx context.Context, questID, userID string) (*QuestWithProgress, error) {
	var quest QuestWithProgress
	err := q.http.Get(ctx, "/quests/"+questID+"/progress/"+url.PathEscape(userID), nil, &quest)
	if err != nil {
		return nil, err
	}
	return &quest, nil
}

// ListWithProgress returns quests with progress for a user
func (q *QuestsClient) ListWithProgress(ctx context.Context, userID string, opts *ListQuestsOptions) ([]QuestWithProgress, error) {
	params := url.Values{}
	params.Set("user_id", userID)
	if opts != nil {
		if opts.Status != "" {
			params.Set("status", opts.Status)
		}
		if opts.Category != "" {
			params.Set("category", opts.Category)
		}
	}

	var quests []QuestWithProgress
	err := q.http.Get(ctx, "/quests/with-progress", params, &quests)
	return quests, err
}

// StartQuest starts a quest for a user
func (q *QuestsClient) StartQuest(ctx context.Context, questID, userID string) (*UserQuestProgress, error) {
	var progress UserQuestProgress
	err := q.http.Post(ctx, "/quests/"+questID+"/start", map[string]interface{}{
		"user_id": userID,
	}, &progress)
	if err != nil {
		return nil, err
	}
	return &progress, nil
}

// GetUserProgress returns user's progress on a quest
func (q *QuestsClient) GetUserProgress(ctx context.Context, questID, userID string) (*UserQuestProgress, error) {
	var progress UserQuestProgress
	err := q.http.Get(ctx, "/quests/"+questID+"/progress/"+url.PathEscape(userID), nil, &progress)
	if err != nil {
		return nil, err
	}
	return &progress, nil
}

// CompleteStep completes a step manually
func (q *QuestsClient) CompleteStep(ctx context.Context, questID, userID, stepID string) (*UserQuestProgress, error) {
	var progress UserQuestProgress
	err := q.http.Post(ctx, "/quests/"+questID+"/steps/"+stepID+"/complete", map[string]interface{}{
		"user_id": userID,
	}, &progress)
	if err != nil {
		return nil, err
	}
	return &progress, nil
}

// GetAllUserProgress returns all quest progress for a user
func (q *QuestsClient) GetAllUserProgress(ctx context.Context, userID string) ([]UserQuestProgress, error) {
	var progress []UserQuestProgress
	err := q.http.Get(ctx, "/quests/user/"+url.PathEscape(userID)+"/progress", nil, &progress)
	return progress, err
}

// AddStep adds a step to a quest
func (q *QuestsClient) AddStep(ctx context.Context, questID string, step *CreateQuestStepRequest) (*QuestStep, error) {
	var result QuestStep
	err := q.http.Post(ctx, "/quests/"+questID+"/steps", step, &result)
	if err != nil {
		return nil, err
	}
	return &result, nil
}

// UpdateStep updates a step
func (q *QuestsClient) UpdateStep(ctx context.Context, questID, stepID string, step *CreateQuestStepRequest) (*QuestStep, error) {
	var result QuestStep
	err := q.http.Put(ctx, "/quests/"+questID+"/steps/"+stepID, step, &result)
	if err != nil {
		return nil, err
	}
	return &result, nil
}

// DeleteStep deletes a step
func (q *QuestsClient) DeleteStep(ctx context.Context, questID, stepID string) error {
	return q.http.Delete(ctx, "/quests/"+questID+"/steps/"+stepID)
}

// ReorderSteps reorders steps
func (q *QuestsClient) ReorderSteps(ctx context.Context, questID string, stepIDs []string) (*Quest, error) {
	var quest Quest
	err := q.http.Post(ctx, "/quests/"+questID+"/steps/reorder", map[string]interface{}{
		"step_ids": stepIDs,
	}, &quest)
	if err != nil {
		return nil, err
	}
	return &quest, nil
}
