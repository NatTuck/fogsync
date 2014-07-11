
window.App = Ember.Application.create {
  rootElement: '.content',
}

skip = ->
  # do nothing

logError = (msg) ->
  console.log("Error:", msg)

App.ApplicationController = Ember.Controller.extend({})

App.Settings = Ember.Object.extend({})
App.Settings.reopenClass({
  findAll: () ->
    $.getJSON('/settings')
  save: () ->
    console.log("TODO: Save settings")
})

App.Share = Ember.Object.extend({})
App.Share.reopenClass({
  findAll: () ->
    $.getJSON('/shares')
  find: (name) ->
    $.getJSON("/shares/#{name}")
})

App.IndexRoute = Ember.Route.extend({
  model: (params) ->
    {}
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

