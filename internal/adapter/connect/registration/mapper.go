package registration

import (
	"strconv"

	devicepb "github.com/quixiq/polyglot/api/gen/v1"
	domainRegistration "github.com/quixiq/polyglot/internal/domain/registration"
)

func toProtoRegistration(r *domainRegistration.Registration) *devicepb.Registration {
	if r == nil {
		return nil
	}
	var lat, lon float64
	hasCoord := false
	if r.Latitude != nil {
		lat = *r.Latitude
		hasCoord = true
	}
	if r.Longitude != nil {
		lon = *r.Longitude
	}
	var schedDateUnix int64
	var schedTime string
	if r.ScheduledInstallDate != nil {
		schedDateUnix = r.ScheduledInstallDate.Unix()
	}
	if r.ScheduledInstallTime != nil {
		schedTime = r.ScheduledInstallTime.Format("15:04")
	}
	var techID string
	if r.AssignedTechnicianID != nil {
		techID = fmtInt(*r.AssignedTechnicianID)
	}
	var installedAtUnix int64
	if r.InstalledAt != nil {
		installedAtUnix = r.InstalledAt.Unix()
	}
	return &devicepb.Registration{
		Id:                       r.ID,
		TenantId:                 r.TenantID,
		RegistrationNo:           r.RegistrationNo,
		PlanId:                   r.PlanID,
		FullName:                 r.FullName,
		Phone:                    r.Phone,
		Email:                    r.Email,
		Address:                  r.Address,
		Latitude:                 lat,
		Longitude:                lon,
		HasCoordinates:           hasCoord,
		Notes:                    r.Notes,
		Status:                   r.Status,
		ScheduledInstallDateUnix: schedDateUnix,
		ScheduledInstallTime:     schedTime,
		AssignedTechnicianId:     techID,
		InstalledAtUnix:          installedAtUnix,
		TechnicianNotes:          r.TechnicianNotes,
		CustomerId:               r.CustomerID,
		SubscriptionId:           r.SubscriptionID,
		InvoiceId:                r.InvoiceID,
		RejectedReason:           r.RejectedReason,
		CancelReason:             r.CancelReason,
		CreatedAtUnix:            r.CreatedAt.Unix(),
	}
}

func toProtoRegistrationList(regs []domainRegistration.Registration) []*devicepb.Registration {
	out := make([]*devicepb.Registration, len(regs))
	for i := range regs {
		out[i] = toProtoRegistration(&regs[i])
	}
	return out
}

func fmtInt(n uint) string {
	return strconv.FormatUint(uint64(n), 10)
}
