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
	var surveyDateUnix int64
	var surveyTime string
	if r.SurveyDate != nil {
		surveyDateUnix = r.SurveyDate.Unix()
	}
	if r.SurveyTime != nil {
		surveyTime = r.SurveyTime.Format("15:04")
	}
	var surveyorID string
	if r.SurveyorID != nil {
		surveyorID = fmtInt(*r.SurveyorID)
	}
	var installationStartedAtUnix int64
	if r.InstallationStartedAt != nil {
		installationStartedAtUnix = r.InstallationStartedAt.Unix()
	}
	return &devicepb.Registration{
		Id:                        r.ID,
		TenantId:                  r.TenantID,
		RegistrationNo:            r.RegistrationNo,
		PlanId:                    r.PlanID,
		FullName:                  r.FullName,
		Phone:                     r.Phone,
		Email:                     r.Email,
		Address:                   r.Address,
		Latitude:                  lat,
		Longitude:                 lon,
		HasCoordinates:            hasCoord,
		Notes:                     r.Notes,
		Status:                    r.Status,
		ScheduledInstallDateUnix:  schedDateUnix,
		ScheduledInstallTime:      schedTime,
		AssignedTechnicianId:      techID,
		TargetDeviceId:            r.TargetDeviceID,
		InstalledAtUnix:           installedAtUnix,
		TechnicianNotes:           r.TechnicianNotes,
		CustomerId:                r.CustomerID,
		SubscriptionId:            r.SubscriptionID,
		InvoiceId:                 r.InvoiceID,
		RejectedReason:            r.RejectedReason,
		CancelReason:              r.CancelReason,
		CreatedAtUnix:             r.CreatedAt.Unix(),
		Nik:                       r.NIK,
		IdCardPhotoUrl:            r.IDCardPhotoURL,
		HousePhotoUrl:             r.HousePhotoURL,
		SurveyDateUnix:            surveyDateUnix,
		SurveyTime:                surveyTime,
		SurveyorId:                surveyorID,
		SurveyStatus:              r.SurveyStatus,
		SurveyNotes:               r.SurveyNotes,
		EstimatedCableLengthMeter: int32(r.EstimatedCableLengthMeter),
		PublicToken:               r.PublicToken,
		InstallationStartedAtUnix: installationStartedAtUnix,
		InstallationPhotoUrl:      r.InstallationPhotoURL,
		OntMacAddress:             r.ONTMACAddress,
		OltDeviceId:               r.OLTDeviceID,
		PonIndex:                  int32PtrToInt32(r.PONIndex),
		OnuIndex:                  int32PtrToInt32(r.ONUIndex),
		OntSerialNumber:           r.ONTSerialNumber,
		DiscountAmount:            r.DiscountAmount,
	}
}

func toProtoRegistrationEvents(events []domainRegistration.RegistrationEvent) []*devicepb.RegistrationEvent {
	out := make([]*devicepb.RegistrationEvent, len(events))
	for i := range events {
		e := events[i]
		var actorID string
		if e.ActorID != nil {
			actorID = fmtInt(*e.ActorID)
		}
		out[i] = &devicepb.RegistrationEvent{
			Id:             e.ID,
			RegistrationId: e.RegistrationID,
			EventType:      e.EventType,
			ActorId:        actorID,
			Title:          e.Title,
			Description:    e.Description,
			MetadataJson:   string(e.Metadata),
			CreatedAtUnix:  e.CreatedAt.Unix(),
		}
	}
	return out
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

func int32PtrToInt32(v *int) int32 {
	if v == nil {
		return 0
	}
	return int32(*v)
}
