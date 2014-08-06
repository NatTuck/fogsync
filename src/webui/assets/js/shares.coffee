
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

App.SharesIndexView = Ember.View.extend({
  templateName: 'shares-index'
  didInsertElement: () ->
    $('.delete-broken').click (ee) ->
      name = $(ee.target).attr('data-name')
      if !confirm("Really delete broken share #{name}?")
        return

      $.ajax "/shares/#{name}",
        type: 'DELETE',
        success: (data, status, xhr) ->
          alert("Deleted #{name}")
          console.log(status)
          console.log(data)
        error: (xhr, status, msg) ->
          alert("Delete Failed: #{msg}")
})


