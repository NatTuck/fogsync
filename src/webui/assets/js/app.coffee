
window.App = Ember.Application.create {
  rootElement: '.content',
}

skip = ->

App.ApplicationController = Ember.Controller.extend({})

App.ApplicationAdapter = DS.RESTAdapter.extend({
  ajaxError: (jqXHR) ->
    error = this._super(jqXHR)
    console.log(error)
})

App.Share = DS.Model.extend({
  Name: DS.attr('string'),
  Root: DS.attr('string'),
  Ckey: DS.attr('string'),
  Hkey: DS.attr('string'),
})

App.Setting = DS.Model.extend({
  Email: DS.attr('string'),
  Cloud: DS.attr('string'),
  Passwd: DS.attr('string'),
  Master: DS.attr('string'),
})

App.IndexRoute = Ember.Route.extend({
  model: (params) ->
    this.store.find('setting', 0)
})

App.IndexController = Ember.Controller.extend({
  actions: {
    save: (ee) ->
      settings = this.get('model')
      settings.save().then(skip, -> console.log(settings.errors))
  },
})

App.SharesRoute = Ember.Route.extend({
  model: (params) ->
    this.store.findAll('share')
})

App.Router.map ->
  this.route("setup")
  this.resource "settings", ->
    this.resource('setting', {path: ':settings_id'})
  this.resource("shares")
  this.route("about")

