
App.Share = Ember.Object.extend({})
App.Share.reopenClass({
  findAll: () ->
    $.getJSON('/shares').then (shares) ->
      shares.map (ss) -> App.Share.create(ss)
  find: (name) ->
    $.getJSON("/shares/#{name}").then (data) ->
      App.Share.create(data)
})

App.SharesRoute = Ember.Route.extend({
  model: (params) ->
    App.Share.findAll()
})

App.ShareRoute = Ember.Route.extend({
  model: (params) ->
    App.Share.find(params['name'])
})

App.SharesIndexView = Ember.View.extend({
  templateName: 'shares-index'
})


