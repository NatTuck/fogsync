
App.Share = Ember.Object.extend({})
App.Share.reopenClass({
  findAll: () ->
    $.getJSON('/shares').then (data) ->
      {
        shares: data.Shares.map (ss) -> App.Share.create(ss)
        broken: data.Broken
      }
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

deleteShare = (name) ->
  $.ajax "/shares/#{name}",
    type: 'DELETE',
    success: (data, status, xhr) ->
      console.log(status)
      console.log(data)
      window.location = "/#/shares"

    error: (xhr, status, msg) ->
      console.log(status)
      console.log(msg)
      alert("Delete Failed: #{msg}")

App.SharesIndexView = Ember.View.extend({
  templateName: 'shares-index'
  didInsertElement: () ->
    $('.delete-broken').click (ee) ->
      name = $(ee.target).attr('data-name')
      if !confirm("Really delete broken share #{name}?")
        return
  
      deleteShare(name)
})

App.ShareView = Ember.View.extend({
  templateName: 'share',
  didInsertElement: () ->
    $('.delete-share').click (ee) ->
      name = $(ee.target).attr('data-name')
      if !confirm("Really delete share #{name}?")
        return

      deleteShare(name)
})
