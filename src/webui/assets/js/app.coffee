
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

App.Settings = Ember.Object.extend({
  dirty: false

  save: () ->
    data = this.getProperties('Email', 'Cloud', 'Passwd', 'Master')
    $.putJSON '/settings', data,  () =>
      this.set('dirty', false)

  changed: (() ->
    this.set('dirty', true)
  ).observes('Email', 'Cloud', 'Passwd', 'Master')
})
App.Settings.reopenClass({
  find: () ->
    $.getJSON('/settings').then (data) ->
      App.Settings.create(data)
})

App.Share = Ember.Object.extend({})
App.Share.reopenClass({
  findAll: () ->
    $.getJSON('/shares').then (shares) ->
      shares.map (ss) -> App.Share.create(ss)
  find: (name) ->
    $.getJSON("/shares/#{name}").then (data) ->
      App.Share.create(data)
})

App.IndexRoute = Ember.Route.extend({
  model: (params) ->
    {}
})

App.SettingsRoute = Ember.Route.extend({
  model: (params) ->
    App.Settings.find()
})

App.IndexController = Ember.Controller.extend({
  actions: {
    save: (ee) ->
      settings = this.get('model')
      settings.save()
  },
})

App.SharesRoute = Ember.Route.extend({
  model: (params) ->
    App.Share.findAll()
})

App.SharesIndexView = Ember.View.extend({
  templateName: 'shares-index'
})

App.ShareRoute = Ember.Route.extend({
  model: (params) ->
    App.Share.find(params['name'])
})

App.Router.map ->
  this.route("setup")
  this.route("settings")
  this.resource "shares", ->
    this.resource('share', {path: ':name'})
  this.route("about")

