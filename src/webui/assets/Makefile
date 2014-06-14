
# Identify coffescript and javascript files.
JS := $(wildcard js/*.js)
CS := $(wildcard js/*.coffee)

# Determine coffee -> js mappings.
CSJS := $(shell echo "$(CS)" | perl -plE 's/^js/public/g; s/coffee\b/js/g')

# Identify css and scss files.
CSS  := $(wildcard css/*.css)
SCSS := $(wildcard css/*.scss)

# Identify handlebars files.
HBS := $(wildcard html/*.hbs)

public/index.html: $(CSJS) public/fogsync.css public/templates.js html/index.html.erb
	cp js/*.* js/bootstrap/* css/*.* public
	erb -T - html/index.html.erb > public/index.html

$(CSJS): $(CS)
	coffee -c -m -o public $<

public/fogsync.css: $(SCSS)
	scss --sourcemap --unix-newlines css/fogsync.scss public/fogsync.css

public/templates.js: $(HBS)
	@echo "/* Here's some handlebars templates */" > public/templates.js
	echo "$(HBS)" | xargs -n 1 ember-precompile >> public/templates.js

clean:
	rm -f public/*
	touch public/.keep

.PHONY: all clean sassy
