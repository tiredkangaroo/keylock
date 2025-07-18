.PHONY: dev server tailwind templ

# Run templ generation in watch mode
templ:
	templ generate --watch

# Run air for Go hot reload
server:
	@echo "Starting Air hot reload..."
	air \
		--build.cmd "go build -o tmp/bin/main ." \
		--build.bin "tmp/bin/main" \
		--build.delay "100" \
		--build.exclude_dir "node_modules" \
		--build.include_ext "go" \
		--build.stop_on_error "false" \
		--misc.clean_on_exit true

# Watch Tailwind CSS changes
tailwind:
	tailwindcss -i ./web/assets/css/input.css -o ./web/assets/css/output.css --watch

# Start development server with all watchers
dev:
	@$(MAKE) tailwind & \
	$(MAKE) templ & \
	$(MAKE) server
