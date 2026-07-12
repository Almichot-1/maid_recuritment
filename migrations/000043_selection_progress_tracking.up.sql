-- Migration: Create selection progress tracking table
-- Replaces the old status_steps system with a comprehensive progress tracking system

CREATE TABLE IF NOT EXISTS selection_progress (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    selection_id UUID NOT NULL REFERENCES selections(id) ON DELETE CASCADE,
    
    -- COC (Certificate of Competency)
    coc_status TEXT NOT NULL DEFAULT 'pending' CHECK (coc_status IN ('pending', 'in_progress', 'done', 'failed')),
    coc_type TEXT CHECK (coc_type IN ('online', 'offline')),
    coc_document_url TEXT,
    coc_document_name TEXT,
    coc_uploaded_at TIMESTAMPTZ,
    
    -- Medical
    medical_status TEXT NOT NULL DEFAULT 'pending' CHECK (medical_status IN ('pending', 'in_progress', 'done', 'failed')),
    medical_document_url TEXT,
    medical_document_name TEXT,
    medical_uploaded_at TIMESTAMPTZ,
    
    -- Visa
    visa_status TEXT NOT NULL DEFAULT 'pending' CHECK (visa_status IN ('pending', 'in_progress', 'approved', 'rejected')),
    visa_document_url TEXT,
    visa_document_name TEXT,
    visa_uploaded_at TIMESTAMPTZ,
    
    -- Ticket
    ticket_status TEXT NOT NULL DEFAULT 'pending' CHECK (ticket_status IN ('pending', 'booked', 'confirmed')),
    ticket_document_url TEXT,
    ticket_document_name TEXT,
    ticket_uploaded_at TIMESTAMPTZ,
    
    -- Arrival
    arrival_status TEXT NOT NULL DEFAULT 'not_arrived' CHECK (arrival_status IN ('not_arrived', 'arrived')),
    arrival_date DATE,
    arrival_city TEXT,
    
    -- Metadata
    updated_by UUID NOT NULL REFERENCES users(id),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    
    -- Ensure one progress record per selection
    CONSTRAINT uq_selection_progress_selection_id UNIQUE(selection_id)
);

-- Indexes for performance
CREATE INDEX idx_selection_progress_selection_id ON selection_progress(selection_id);
CREATE INDEX idx_selection_progress_updated_by ON selection_progress(updated_by);
CREATE INDEX idx_selection_progress_coc_status ON selection_progress(coc_status);
CREATE INDEX idx_selection_progress_medical_status ON selection_progress(medical_status);
CREATE INDEX idx_selection_progress_visa_status ON selection_progress(visa_status);
CREATE INDEX idx_selection_progress_ticket_status ON selection_progress(ticket_status);
CREATE INDEX idx_selection_progress_arrival_status ON selection_progress(arrival_status);

-- Add comment
COMMENT ON TABLE selection_progress IS 'Tracks the progress of candidate recruitment process per selection (COC, Medical, Visa, Ticket, Arrival)';
