
# Identify coffescript and javascript files.
JS := $(wildcard js/*.js)
CS := $(wildcard js/*.coffee)

# Determine coffee -> js mappings.
CSJS := $(shell echo "$(CS)" | perl -plE 's/^js/public/; s/coffee\b/js/')

# Identify css and scss files.
CSS  := $(wildcard css/*.css)
SCSS := $(wildcard css/*.scss)

# Determine scss -> css mappings.
GCSS := $(shell echo "$(SCSS)" | perl -plE 's/^css/public/; s/scss\b/css/')

# Identify handlebars files.
HBS := $(wildcard html/*.hbs)

# Determine hbs > js mapping.
HJS := $(shell echo "$(HBS)" | perl -plE 's/^html/tmp/; s/hbs\b/js/')

public/index.html: $(CSJS) $(GCSS) public/templates.js html/index.html.erb
	cp js/* css/* public
	echo public/*.js > tmp/javascripts.txt
	erb -T - html/index.html.erb > public/index.html

$(CSJS): $(CS)
	coffee -c -m -o public $<

$(GCSS): $(SCSS)
	scss --unix-newlines --sourcemap $< $@

public/templates.js: $(HJS)
	cat tmp/*.js > public/templates.js

$(HJS): $(HBS)
	handlebars $< > $@

clean:
	rm -f public/*
	touch public/.keep
	rm -f tmp/*
	touch tmp/.keep

.PHONY: all clean
