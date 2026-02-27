package repository

import (
    "context"
    "time"

    "github.com/heyjasonn/ai-workflow/internal/workflow"
    "gorm.io/driver/postgres"
    "gorm.io/gorm"
)

type WorkflowModel struct {
    ID                  string `gorm:"primaryKey"`
    Requirement         string
    State               string
    ResearchJSON        []byte `gorm:"type:jsonb"`
    PlanJSON            []byte `gorm:"type:jsonb"`
    TaskListJSON        []byte `gorm:"type:jsonb"`
    ExecutionResultJSON []byte `gorm:"type:jsonb"`
    TestReportJSON      []byte `gorm:"type:jsonb"`
    ApprovalsJSON       []byte `gorm:"type:jsonb"`
    CreatedAt           time.Time
}

func (WorkflowModel) TableName() string {
    return "workflows"
}

type GormRepository struct {
    db *gorm.DB
}

func OpenGormPostgres(dsn string) (*gorm.DB, error) {
    return gorm.Open(postgres.Open(dsn), &gorm.Config{})
}

func NewGormRepository(db *gorm.DB) *GormRepository {
    return &GormRepository{db: db}
}

func (r *GormRepository) EnsureSchema(ctx context.Context) error {
    return r.db.WithContext(ctx).AutoMigrate(&WorkflowModel{})
}

func (r *GormRepository) Create(ctx context.Context, wf *workflow.Workflow) error {
    model := toModel(wf)
    return r.db.WithContext(ctx).Create(&model).Error
}

func (r *GormRepository) Get(ctx context.Context, id string) (*workflow.Workflow, error) {
    var model WorkflowModel
    err := r.db.WithContext(ctx).First(&model, "id = ?", id).Error
    if err != nil {
        if err == gorm.ErrRecordNotFound {
            return nil, workflow.ErrNotFound
        }
        return nil, err
    }
    wf := fromModel(&model)
    return wf, nil
}

func (r *GormRepository) Update(ctx context.Context, wf *workflow.Workflow) error {
    model := toModel(wf)
    tx := r.db.WithContext(ctx).Model(&WorkflowModel{}).Where("id = ?", wf.ID).Updates(model)
    if tx.Error != nil {
        return tx.Error
    }
    if tx.RowsAffected == 0 {
        return workflow.ErrNotFound
    }
    return nil
}

func (r *GormRepository) List(ctx context.Context) ([]*workflow.Workflow, error) {
    var models []WorkflowModel
    if err := r.db.WithContext(ctx).Order("created_at desc").Find(&models).Error; err != nil {
        return nil, err
    }
    items := make([]*workflow.Workflow, 0, len(models))
    for i := range models {
        items = append(items, fromModel(&models[i]))
    }
    return items, nil
}

func toModel(wf *workflow.Workflow) WorkflowModel {
    return WorkflowModel{
        ID:                  wf.ID,
        Requirement:         wf.Requirement,
        State:               string(wf.State),
        ResearchJSON:        wf.ResearchJSON,
        PlanJSON:            wf.PlanJSON,
        TaskListJSON:        wf.TaskListJSON,
        ExecutionResultJSON: wf.ExecutionResultJSON,
        TestReportJSON:      wf.TestReportJSON,
        ApprovalsJSON:       wf.ApprovalsJSON,
        CreatedAt:           wf.CreatedAt,
    }
}

func fromModel(model *WorkflowModel) *workflow.Workflow {
    return &workflow.Workflow{
        ID:                  model.ID,
        Requirement:         model.Requirement,
        State:               workflow.WorkflowState(model.State),
        ResearchJSON:        model.ResearchJSON,
        PlanJSON:            model.PlanJSON,
        TaskListJSON:        model.TaskListJSON,
        ExecutionResultJSON: model.ExecutionResultJSON,
        TestReportJSON:      model.TestReportJSON,
        ApprovalsJSON:       model.ApprovalsJSON,
        CreatedAt:           model.CreatedAt,
    }
}
