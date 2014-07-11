
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

App.SettingsRoute = Ember.Route.extend({
  model: (params) ->
    App.Settings.find()
})

App.SettingsController = Ember.Controller.extend({
  actions: {
    save: (ee) ->
      settings = this.get('model')
      settings.save()
  },
})


