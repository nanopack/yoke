###
# Compass
###

# Change Compass configuration
# compass_config do |config|
#   config.output_style = :compact
# end

###
# Page options, layouts, aliases and proxies
###

# Per-page layout changes:
#
# With no layout
# page "/path/to/file.html", :layout => false
#
# With alternative layout
# page "/path/to/file.html", :layout => :otherlayout
#
# A path which all have the same layout
# with_layout :admin do
#   page "/admin/*"
# end

# Proxy pages (https://middlemanapp.com/advanced/dynamic_pages/)
# proxy "/this-page-has-no-template.html", "/template-file.html", :locals => {
#  :which_fake_page => "Rendering a fake page with a local variable" }

# Using Github Flavored Markdown
set :markdown_engine, :redcarpet
set :markdown, :tables => true, :autolink => true, :gh_blockcode => true, :fenced_code_blocks => true, :strikethrough => true, :with_toc_data => true

activate :directory_indexes

###
# Helpers
###

# Automatic image dimensions on image_tag helper
# activate :automatic_image_sizes

# Reload the browser automatically whenever files change
configure :development do
  activate :livereload do |opts|
    opts.host = '0.0.0.0'
    opts.js_host = 'yolk-site.dev'
  end
end


# Methods defined in the helpers block are available in templates
# Methods defined in the helpers block are available in templates
helpers do
  def nav_article_active(path)
    current_page.url.gsub(/\A\//, '') === path ? {:class => "active"} : {}
  end
end

set :css_dir, 'stylesheets'

set :js_dir, 'javascripts'

set :images_dir, 'images'

set :partials_dir, 'partials'

# Used when building links in the left nav of documentation
set :nav_root, '/'

# Build-specific configuration
configure :build do
  # For example, change the Compass output style for deployment
  # activate :minify_css

  # Minify Javascript on build
  # activate :minify_javascript

  # Enable cache buster
  # activate :asset_hash

  # Use relative URLs
  # activate :relative_assets

  # Or use a different image path
  # set :http_prefix, "/Content/images/"

  # Used when building links in the left nav of documentation
  set :nav_root, '/' + data.project_info.name.downcase.gsub(/\s+/, "") + '/'
end
