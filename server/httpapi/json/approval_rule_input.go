package json

import (
	"encoding/json"
	"errors"

	"github.com/fullstaq-labs/sqedule/server/dbmodels"
	"github.com/fullstaq-labs/sqedule/server/dbmodels/approvalpolicy"
	"github.com/fullstaq-labs/sqedule/server/dbmodels/retrypolicy"
)

//
// ******** Types, constants & variables ********/
//

type ApprovalRuleInput struct {
	Type dbmodels.ApprovalRuleType
	ApprovalRuleInputBase
	HTTPApiApprovalRuleInput
	ScheduleApprovalRuleInput
	ManualApprovalRuleInput
}

type ApprovalRuleInputBase struct {
	Enabled *bool `json:"enabled"`
}

type HTTPApiApprovalRuleInput struct {
	URL              string             `json:"url"`
	Username         *string            `json:"username"`
	Password         *string            `json:"password"`
	TLSCaCertificate *string            `json:"tls_ca_certificate"`
	RetryPolicy      retrypolicy.Policy `json:"retry_policy"`
	RetryLimit       int                `json:"retry_limit"`
}

type ScheduleApprovalRuleInput struct {
	BeginTime    *string `json:"begin_time"`
	EndTime      *string `json:"end_time"`
	DaysOfWeek   *string `json:"days_of_week"`
	DaysOfMonth  *string `json:"days_of_month"`
	MonthsOfYear *string `json:"months_of_year"`
}

type ManualApprovalRuleInput struct {
	ApprovalPolicy approvalpolicy.Policy `json:"approval_policy"`
	Minimum        *int32                `json:"minimum"`
}

//
// ******** ApprovalRuleInput methods ********/
//

func (input *ApprovalRuleInput) UnmarshalJSON(b []byte) error {
	var object map[string]interface{}
	err := json.Unmarshal(b, &object)
	if err != nil {
		return err
	}

	if object["type"] == nil {
		return errors.New("Unspecified approval rule type")
	}

	input.Type = dbmodels.ApprovalRuleType(object["type"].(string))
	switch input.Type {
	case dbmodels.HTTPApiApprovalRuleType:
		return json.Unmarshal(b, &input.HTTPApiApprovalRuleInput)
	case dbmodels.ScheduleApprovalRuleType:
		return json.Unmarshal(b, &input.ScheduleApprovalRuleInput)
	case dbmodels.ManualApprovalRuleType:
		return json.Unmarshal(b, &input.ManualApprovalRuleInput)
	default:
		panic("Unsupported approval rule type " + input.Type)
	}
}

func (input ApprovalRuleInput) AppendToDbmodelsApprovalRulesetContents(organizationID string, contents *dbmodels.ApprovalRulesetContents) {
	base := dbmodels.ApprovalRule{
		BaseModel: dbmodels.BaseModel{
			OrganizationID: organizationID,
		},
	}

	switch input.Type {
	case dbmodels.HTTPApiApprovalRuleType:
		model := dbmodels.HTTPApiApprovalRule{ApprovalRule: base}
		input.ApprovalRuleInputBase.PopulateDbmodel(&model.ApprovalRule)
		input.HTTPApiApprovalRuleInput.PopulateDbmodel(&model)
		contents.HTTPApiApprovalRules = append(contents.HTTPApiApprovalRules, model)

	case dbmodels.ScheduleApprovalRuleType:
		model := dbmodels.ScheduleApprovalRule{ApprovalRule: base}
		input.ApprovalRuleInputBase.PopulateDbmodel(&model.ApprovalRule)
		input.ScheduleApprovalRuleInput.PopulateDbmodel(&model)
		contents.ScheduleApprovalRules = append(contents.ScheduleApprovalRules, model)

	case dbmodels.ManualApprovalRuleType:
		model := dbmodels.ManualApprovalRule{ApprovalRule: base}
		input.ApprovalRuleInputBase.PopulateDbmodel(&model.ApprovalRule)
		input.ManualApprovalRuleInput.PopulateDbmodel(&model)
		contents.ManualApprovalRules = append(contents.ManualApprovalRules, model)

	default:
		panic("Unsupported approval rule type " + input.Type)
	}
}

//
// ******** ApprovalRuleInputBase methods ********/
//

func (input ApprovalRuleInputBase) PopulateDbmodel(model *dbmodels.ApprovalRule) {
	model.Enabled = input.Enabled
}

//
// ******** HTTPApiApprovalRuleInput methods ********/
//

func (input HTTPApiApprovalRuleInput) PopulateDbmodel(model *dbmodels.HTTPApiApprovalRule) {
	model.URL = input.URL
	model.Username = stringPointerToSqlString(input.Username)
	model.Password = stringPointerToSqlString(input.Password)
	model.TLSCaCertificate = stringPointerToSqlString(input.TLSCaCertificate)
	model.RetryPolicy = input.RetryPolicy
	model.RetryLimit = input.RetryLimit
}

//
// ******** ScheduleApprovalRuleInput methods ********/
//

func (input ScheduleApprovalRuleInput) PopulateDbmodel(model *dbmodels.ScheduleApprovalRule) {
	model.BeginTime = stringPointerToSqlString(input.BeginTime)
	model.EndTime = stringPointerToSqlString(input.EndTime)
	model.DaysOfWeek = stringPointerToSqlString(input.DaysOfWeek)
	model.DaysOfMonth = stringPointerToSqlString(input.DaysOfMonth)
	model.MonthsOfYear = stringPointerToSqlString(input.MonthsOfYear)
}

//
// ******** ManualApprovalRuleInput methods ********/
//

func (input ManualApprovalRuleInput) PopulateDbmodel(model *dbmodels.ManualApprovalRule) {
	model.ApprovalPolicy = input.ApprovalPolicy
	model.Minimum = int32PointerToSqlInt32(input.Minimum)
}
