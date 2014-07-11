
window.App = Ember.Application.create {
  rootElement: '.content',
}

skip = ->
  # do nothing

logError = (msg) ->
  console.log("Error:", msg)

$.extend({
	putJSON: (url, data, callback) ->
		$.ajax({
			type: 'put',
			url: url,
			processData: false,
			data: JSON.stringify(data, null, 2),
			success: callback,
			contentType: 'application/json',
			dataType: 'json'
    })
})

App.ApplicationController = Ember.Controller.extend({})

App.IndexRoute = Ember.Route.extend({
  model: (params) ->
    {}
})

App.IndexController = Ember.Controller.extend({})

App.Router.map ->
  this.route("setup")
  this.route("settings")
  this.resource "shares", ->
    this.resource('share', {path: ':name'})
  this.route("about")

