CREATE TABLE IF NOT EXISTS platform_settings (
    id text PRIMARY KEY,
    selection_lock_duration_hours integer NOT NULL DEFAULT 24,
    require_both_approvals boolean NOT NULL DEFAULT true,
    auto_approve_agencies boolean NOT NULL DEFAULT false,
    auto_expire_selections boolean NOT NULL DEFAULT true,
    email_notifications_enabled boolean NOT NULL DEFAULT true,
    maintenance_mode boolean NOT NULL DEFAULT false,
    maintenance_message text NOT NULL DEFAULT '',
    agency_approval_email_template text NOT NULL DEFAULT '',
    agency_rejection_email_template text NOT NULL DEFAULT '',
    selection_notification_email_template text NOT NULL DEFAULT '',
    expiry_notification_email_template text NOT NULL DEFAULT '',
    created_at timestamptz NOT NULL DEFAULT now(),
    updated_at timestamptz NOT NULL DEFAULT now()
);

INSERT INTO platform_settings (
    id,
    selection_lock_duration_hours,
    require_both_approvals,
    auto_approve_agencies,
    auto_expire_selections,
    email_notifications_enabled,
    maintenance_mode,
    maintenance_message,
    agency_approval_email_template,
    agency_rejection_email_template,
    selection_notification_email_template,
    expiry_notification_email_template
)
VALUES (
    'default',
    24,
    true,
    false,
    true,
    true,
    false,
    'The platform is currently under scheduled maintenance. Please try again later.',
    'Hello {full_name},\n\nYour account for {company_name} has been approved. You can now log in to the platform.',
    'Hello {full_name},\n\nYour application for {company_name} was rejected.\nReason: {reason}',
    'Hello {full_name},\n\n{message}\n\nCandidate: {candidate_name}',
    'Hello {full_name},\n\n{message}\n\nCandidate: {candidate_name}'
)
ON CONFLICT (id) DO NOTHING;

DO $$
BEGIN
    IF NOT EXISTS (
        SELECT 1
        FROM pg_trigger
        WHERE tgname = 'set_platform_settings_updated_at'
    ) THEN
        CREATE TRIGGER set_platform_settings_updated_at
        BEFORE UPDATE ON platform_settings
        FOR EACH ROW
        EXECUTE FUNCTION set_updated_at();
    END IF;
END
$$;
