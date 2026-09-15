-- Revert: Remove oauth2_flavor column from thirdparty_oauth2_services table

ALTER TABLE thirdparty_oauth2_services
    DROP COLUMN oauth2_flavor;
