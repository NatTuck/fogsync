
window.App = Ember.Application.create {
  rootElement: '.content',
}

App.ApplicationController = Ember.Controller.extend({})

App.ApplicationView = Ember.View.extend({
  templateName: 'application',
})

App.Router.map ->
  this.route("about")
  this.route("contact")
