
App.About = Ember.Object.extend({})

App.About.reopenClass({
  find: () ->
    $.getJSON('/about').then (data) ->
      App.Settings.create(data)
})

App.AboutRoute = Ember.Route.extend({
  model: (params) ->
    App.About.find()
})

