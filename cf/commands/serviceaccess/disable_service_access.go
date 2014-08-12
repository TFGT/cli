package serviceaccess

import (
	"github.com/cloudfoundry/cli/cf/actors"
	"github.com/cloudfoundry/cli/cf/command_metadata"
	"github.com/cloudfoundry/cli/cf/configuration"
	"github.com/cloudfoundry/cli/cf/flag_helpers"
	. "github.com/cloudfoundry/cli/cf/i18n"
	"github.com/cloudfoundry/cli/cf/requirements"
	"github.com/cloudfoundry/cli/cf/terminal"
	"github.com/codegangsta/cli"
)

type DisableServiceAccess struct {
	ui     terminal.UI
	config configuration.Reader
	actor  actors.ServicePlanActor
}

func NewDisableServiceAccess(ui terminal.UI, config configuration.Reader, actor actors.ServicePlanActor) (cmd *DisableServiceAccess) {
	return &DisableServiceAccess{
		ui:     ui,
		config: config,
		actor:  actor,
	}
}

func (cmd *DisableServiceAccess) GetRequirements(requirementsFactory requirements.Factory, context *cli.Context) ([]requirements.Requirement, error) {
	if len(context.Args()) != 1 {
		cmd.ui.FailWithUsage(context)
	}

	return []requirements.Requirement{requirementsFactory.NewLoginRequirement()}, nil
}

func (cmd *DisableServiceAccess) Metadata() command_metadata.CommandMetadata {
	return command_metadata.CommandMetadata{
		Name:        "disable-service-access",
		Description: T("Disable access to a service or service plan for one or all orgs"),
		Usage:       "CF_NAME disable-service-access SERVICE [-p PLAN] [-o ORG]",
		Flags: []cli.Flag{
			flag_helpers.NewStringFlag("p", T("Disable access to a particular service plan")),
			flag_helpers.NewStringFlag("o", T("Disable access to a particular organization")),
		},
	}
}

func (cmd *DisableServiceAccess) Run(c *cli.Context) {
	serviceName := c.Args()[0]
	planName := c.String("p")
	orgName := c.String("o")

	if planName != "" && orgName != "" {
		cmd.DisablePlanAndOrgForService(serviceName, planName, orgName)
	} else if planName != "" {
		cmd.DisableSinglePlanForService(serviceName, planName)
	} else if orgName != "" {
		cmd.disablePlansForSingleOrgForService(serviceName, orgName)
	} else {
		cmd.DisableServiceForAll(serviceName)
	}

	cmd.ui.Say(T("OK"))
}

func (cmd *DisableServiceAccess) DisableServiceForAll(serviceName string) {
	allPlansAlreadySet, err := cmd.actor.UpdateAllPlansForService(serviceName, false)
	if err != nil {
		cmd.ui.Failed(err.Error())
	}

	if allPlansAlreadySet {
		cmd.ui.Say(T("All plans of the service are already inaccessible for all orgs"))
	} else {
		cmd.ui.Say(T("Disabling access to all plans of service {{.ServiceName}} for all orgs as {{.UserName}}...", map[string]interface{}{"ServiceName": terminal.EntityNameColor(serviceName), "UserName": terminal.EntityNameColor(cmd.config.Username())}))
	}
}

func (cmd *DisableServiceAccess) DisablePlanAndOrgForService(serviceName string, planName string, orgName string) {
	planOriginalAccess, err := cmd.actor.UpdatePlanAndOrgForService(serviceName, planName, orgName, false)
	if err != nil {
		cmd.ui.Failed(err.Error())
	}

	if planOriginalAccess == actors.None {
		cmd.ui.Say(T("The plan {{.PlanName}} of service {{.ServiceName}} is already accessible for all orgs and no action has been taken at this time.", map[string]interface{}{"PlanName": terminal.EntityNameColor(planName), "ServiceName": terminal.EntityNameColor(serviceName)}))
	} else if planOriginalAccess == actors.Limited {
		cmd.ui.Say(T("Disabling access to plan {{.PlanName}} of service {{.ServiceName}} for org {{.OrgName}} as {{.Username}}...", map[string]interface{}{"PlanName": terminal.EntityNameColor(planName), "ServiceName": terminal.EntityNameColor(serviceName), "OrgName": terminal.EntityNameColor(orgName), "Username": terminal.EntityNameColor(cmd.config.Username())}))

	} else {
		cmd.ui.Say(T("The plan {{.PlanName}} of service {{.ServiceName}} is already inaccessible for org {{.OrgName}}", map[string]interface{}{"PlanName": terminal.EntityNameColor(planName), "ServiceName": terminal.EntityNameColor(serviceName), "OrgName": terminal.EntityNameColor(orgName)}))
	}
	return
}

func (cmd *DisableServiceAccess) DisableSinglePlanForService(serviceName string, planName string) {
	planOriginalAccess, err := cmd.actor.UpdateSinglePlanForService(serviceName, planName, false)
	if err != nil {
		cmd.ui.Failed(err.Error())
	}

	if planOriginalAccess == actors.None {
		cmd.ui.Say(T("The plan {{.PlanName}} of service {{.ServiceName}} is already inaccessible for all orgs", map[string]interface{}{"PlanName": terminal.EntityNameColor(planName), "ServiceName": terminal.EntityNameColor(serviceName)}))
	} else {
		cmd.ui.Say(T("Disabling access of plan {{.PlanName}} for service {{.ServiceName}} as {{.Username}}...", map[string]interface{}{"PlanName": terminal.EntityNameColor(planName), "ServiceName": terminal.EntityNameColor(serviceName), "Username": terminal.EntityNameColor(cmd.config.Username())}))
	}
	return
}

func (cmd *DisableServiceAccess) disablePlansForSingleOrgForService(serviceName string, orgName string) {
	serviceAccess, err := cmd.actor.FindServiceAccess(serviceName)
	if err != nil {
		cmd.ui.Failed(err.Error())
	}
	if serviceAccess == actors.AllPlansArePublic {
		cmd.ui.Say(T("Plans are accessible for all orgs. Try removing access for all orgs, then enable access for select orgs."))
		return
	}

	allPlansWereSet, err := cmd.actor.UpdateOrgForService(serviceName, orgName, false)
	if err != nil {
		cmd.ui.Failed(err.Error())
	}

	if allPlansWereSet {
		cmd.ui.Say(T("All plans of the service are already inaccessible for the org"))
	} else {
		cmd.ui.Say(T("Disabling access to all plans of service {{.ServiceName}} for the org {{.OrgName}} as {{.Username}}...", map[string]interface{}{"ServiceName": terminal.EntityNameColor(serviceName), "OrgName": terminal.EntityNameColor(orgName), "Username": terminal.EntityNameColor(cmd.config.Username())}))
	}
}