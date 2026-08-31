package provisioner

import (
	"context"
	"fmt"
	"strings"

	domainSub "github.com/quixiq/polyglot/internal/domain/subscription"
	"github.com/quixiq/polyglot/internal/port"
	"github.com/quixiq/polyglot/pkg/logger"
)

const dedicatedQueuePrefix = "dq-"

func dedicatedQueueName(username string) string {
	return dedicatedQueuePrefix + username
}

func isDedicated(serviceType string) bool {
	return strings.EqualFold(serviceType, "DEDICATED")
}

// dedicatedQueueFromAccount builds simple queue parameters from subscriber account.
// LimitAt = base rate limit (guaranteed CIR); MaxLimit = full rate limit (including burst segments).
func dedicatedQueueFromAccount(a port.SubscriberAccount, comment string) port.DedicatedQueueParams {
	p := port.DedicatedQueueParams{
		Name:     dedicatedQueueName(a.Username),
		Target:   a.Username,
		MaxLimit: a.RateLimit,
		LimitAt:  a.BaseRateLimit,
		Comment:  comment,
	}
	seg := strings.Split(p.MaxLimit, "/")
	if len(seg) >= 4 {
		p.BurstLimit = seg[2] + "/" + seg[3]
	}
	if len(seg) >= 6 {
		p.BurstThreshold = seg[4] + "/" + seg[5]
	}
	if len(seg) >= 8 {
		p.BurstTime = seg[6] + "/" + seg[7]
	}
	return p
}

// ProvisionDedicated provisions a dedicated subscriber (PPPoE secret + Simple Queue for CIR).
func (p *Provisioner) ProvisionDedicated(ctx context.Context, deviceID string, spec domainSub.DedicatedProvisionSpec) error {
	if err := p.ProvisionPPPoE(ctx, deviceID, spec.PPPoE); err != nil {
		return err
	}
	driver, err := p.resolve(ctx, deviceID)
	if err != nil {
		return fmt.Errorf("resolve driver %s: %w", deviceID, err)
	}
	if p.q != nil && spec.Queue.QueueName != "" {
		qParams := port.DedicatedQueueParams{
			Name:     spec.Queue.QueueName,
			Target:   spec.Queue.Target,
			MaxLimit: spec.Queue.MaxLimit,
			LimitAt:  spec.Queue.LimitAt,
			Comment:  spec.Queue.Comment,
		}
		if _, err := p.q.AddQueue(ctx, driver, qParams); err != nil {
			return fmt.Errorf("add dedicated queue %s: %w", spec.Queue.QueueName, err)
		}
	}
	return nil
}

func (p *Provisioner) ensureDedicatedQueue(ctx context.Context, driver port.DeviceDriver, acct port.SubscriberAccount) error {
	if p.q == nil || acct.Username == "" || acct.RateLimit == "" {
		return nil
	}
	params := dedicatedQueueFromAccount(acct, "AUTO dedicated queue "+acct.Comment)
	name := params.Name

	existing, err := p.q.ListQueues(ctx, driver, name)
	if err != nil {
		return fmt.Errorf("list queues: %w", err)
	}
	for _, qy := range existing {
		if qy.Name != name {
			continue
		}
		if qy.MaxLimit == params.MaxLimit && qy.LimitAt == params.LimitAt {
			return nil
		}
		if _, err := p.q.RemoveQueue(ctx, driver, qy.RosID); err != nil {
			return fmt.Errorf("remove stale queue %s: %w", name, err)
		}
		break
	}
	if _, err := p.q.AddQueue(ctx, driver, params); err != nil {
		return fmt.Errorf("add queue %s: %w", name, err)
	}
	return nil
}

func (p *Provisioner) setDedicatedQueueEnabled(ctx context.Context, driver port.DeviceDriver, username string, enabled bool) {
	p.withDedicatedQueue(ctx, driver, username, func(qy port.SimpleQueue) error {
		_, err := p.q.SetQueueEnabled(ctx, driver, qy.RosID, enabled)
		return err
	})
}

func (p *Provisioner) removeDedicatedQueue(ctx context.Context, driver port.DeviceDriver, username string) {
	p.withDedicatedQueue(ctx, driver, username, func(qy port.SimpleQueue) error {
		_, err := p.q.RemoveQueue(ctx, driver, qy.RosID)
		return err
	})
}

func (p *Provisioner) withDedicatedQueue(ctx context.Context, driver port.DeviceDriver, username string, action func(port.SimpleQueue) error) {
	if p.q == nil || username == "" {
		return
	}
	queues, err := p.q.ListQueues(ctx, driver, dedicatedQueueName(username))
	if err != nil {
		logger.WithComponent("Provisioner").WithFields(map[string]any{
			"username": username,
		}).WithError(err).Warn("list dedicated queue failed")
		return
	}
	for _, qy := range queues {
		if qy.Name == dedicatedQueueName(username) {
			if err := action(qy); err != nil {
				logger.WithComponent("Provisioner").WithFields(map[string]any{
					"username": username,
					"queue":    qy.Name,
				}).WithError(err).Warn("dedicated queue action failed")
			}
			return
		}
	}
}
