package auth

import (
	apiv1connect "github.com/jamesread/starapp/service/gen/starapp/api/v1/apiv1connect"
	"github.com/jamesread/starapp/service/internal/rbac"
)

func RequiredPermission(procedureName string) string {
	switch procedureName {
	case apiv1connect.StarAppServiceGetUserProcedure:
		return rbac.PermissionUsersView
	case apiv1connect.StarAppServiceCreateUserProcedure:
		return rbac.PermissionUsersCreate
	case apiv1connect.StarAppServiceDeleteUserProcedure:
		return rbac.PermissionUsersDelete
	case apiv1connect.StarAppServiceResetUserPasswordProcedure:
		return rbac.PermissionUsersResetPassword

	case apiv1connect.StarAppServiceListRbacPermissionsProcedure,
		apiv1connect.StarAppServiceListRbacRolesProcedure,
		apiv1connect.StarAppServiceGetUserRbacRolesProcedure,
		apiv1connect.StarAppServiceGetUserGroupRbacRolesProcedure,
		apiv1connect.StarAppServiceGetRbacRoleUsersProcedure,
		apiv1connect.StarAppServiceGetRbacRoleGroupsProcedure:
		return rbac.PermissionRbacView

	case apiv1connect.StarAppServiceCreateRbacRoleProcedure,
		apiv1connect.StarAppServiceUpdateRbacRoleProcedure,
		apiv1connect.StarAppServiceDeleteRbacRoleProcedure,
		apiv1connect.StarAppServiceSetUserGroupRbacRolesProcedure:
		return rbac.PermissionRbacManage

	case apiv1connect.StarAppServiceGetUserGroupMembersProcedure:
		return rbac.PermissionUserGroupsView

	case apiv1connect.StarAppServiceCreateUserGroupProcedure,
		apiv1connect.StarAppServiceDeleteUserGroupProcedure,
		apiv1connect.StarAppServiceSetUserGroupMembersProcedure:
		return rbac.PermissionUserGroupsManage

	case apiv1connect.StarAppServiceListCvarsProcedure,
		apiv1connect.StarAppServiceUpdateCvarProcedure,
		apiv1connect.StarAppServiceListWebhooksProcedure,
		apiv1connect.StarAppServiceCreateWebhookProcedure,
		apiv1connect.StarAppServiceUpdateWebhookProcedure,
		apiv1connect.StarAppServiceDeleteWebhookProcedure,
		apiv1connect.StarAppServiceListWebhookDeliveriesProcedure,
		apiv1connect.StarAppServiceFireTestWebhooksProcedure:
		return rbac.PermissionSystemSettings

	case apiv1connect.StarAppServiceImpersonateUserProcedure:
		return rbac.PermissionSystemImpersonate

	case apiv1connect.StarAppServiceStopImpersonationProcedure:
		return ""

	case apiv1connect.StarAppServiceGetMyFamilyProcedure,
		apiv1connect.StarAppServiceListMembersProcedure:
		return rbac.PermissionFamilyView

	case apiv1connect.StarAppServiceCreateFamilyProcedure:
		return rbac.PermissionFamilyManage

	case apiv1connect.StarAppServiceCreateChildMemberProcedure,
		apiv1connect.StarAppServiceUpdateMemberProcedure,
		apiv1connect.StarAppServiceDeleteMemberProcedure:
		return rbac.PermissionMembersManage

	case apiv1connect.StarAppServiceUploadMemberAvatarProcedure,
		apiv1connect.StarAppServiceDeleteMemberAvatarProcedure,
		apiv1connect.StarAppServiceListMemberAvatarsProcedure,
		apiv1connect.StarAppServiceSelectMemberAvatarProcedure:
		return rbac.PermissionMembersAvatar

	case apiv1connect.StarAppServiceAwardStarsProcedure:
		return rbac.PermissionStarsAward

	case apiv1connect.StarAppServiceRevokeStarsProcedure:
		return rbac.PermissionStarsRevoke

	case apiv1connect.StarAppServiceGetMemberBalanceProcedure,
		apiv1connect.StarAppServiceListLedgerProcedure:
		return rbac.PermissionAppAccess

	case apiv1connect.StarAppServiceListRewardsProcedure:
		return rbac.PermissionRewardsView

	case apiv1connect.StarAppServiceCreateRewardProcedure,
		apiv1connect.StarAppServiceUpdateRewardProcedure,
		apiv1connect.StarAppServiceDeleteRewardProcedure:
		return rbac.PermissionRewardsManage

	case apiv1connect.StarAppServiceRequestRedemptionProcedure:
		return rbac.PermissionRedemptionsRequest

	case apiv1connect.StarAppServiceApproveRedemptionProcedure,
		apiv1connect.StarAppServiceRejectRedemptionProcedure,
		apiv1connect.StarAppServiceListRedemptionsProcedure:
		return rbac.PermissionRedemptionsApprove

	case apiv1connect.StarAppServiceGetParentHomeSummaryProcedure:
		return rbac.PermissionStarsViewFamily

	case apiv1connect.StarAppServiceGetChildHomeSummaryProcedure:
		return rbac.PermissionStarsViewOwn

	case apiv1connect.StarAppServiceListChoresProcedure,
		apiv1connect.StarAppServiceListChorePausesProcedure,
		apiv1connect.StarAppServiceGetWeeklyStarChartProcedure,
		apiv1connect.StarAppServiceListStarChartsProcedure:
		return rbac.PermissionChoresViewFamily

	case apiv1connect.StarAppServiceCreateChoreProcedure,
		apiv1connect.StarAppServiceUpdateChoreProcedure,
		apiv1connect.StarAppServiceDeleteChoreProcedure,
		apiv1connect.StarAppServiceCreateChorePauseProcedure,
		apiv1connect.StarAppServiceDeleteChorePauseProcedure,
		apiv1connect.StarAppServiceCreateStarChartProcedure,
		apiv1connect.StarAppServiceUpdateStarChartProcedure,
		apiv1connect.StarAppServiceDeleteStarChartProcedure:
		return rbac.PermissionChoresManage

	case apiv1connect.StarAppServiceCompleteChoreProcedure,
		apiv1connect.StarAppServiceUncompleteChoreProcedure:
		return rbac.PermissionChoresComplete
	}
	return rbac.PermissionAppAccess
}
