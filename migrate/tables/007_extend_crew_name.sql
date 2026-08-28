--liquibase formatted sql

--changeset holeyko:extend_crew_name_to_128
ALTER TABLE IF EXISTS crews
    ALTER COLUMN name TYPE VARCHAR(128);
