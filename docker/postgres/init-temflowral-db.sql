-- Create the application database alongside Temporal's databases on first boot.
SELECT 'CREATE DATABASE temflowral'
WHERE NOT EXISTS (SELECT FROM pg_database WHERE datname = 'temflowral')\gexec
