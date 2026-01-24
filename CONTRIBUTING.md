# Contributing

## DB Migrations

To add a database migration, create a new file in `migrations`. Increment the counter in the filename. Use only snake_case file names. This project only has `up` migrations, so you'll end up with something like `004_my_migration_name.up.sql`.

When the app builds, these are bundled into the executable. When the app is started it will attempt to run the migrations. Test your migration by building the app.
