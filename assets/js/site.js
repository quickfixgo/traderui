$(function() {
  var form = $('#order-ticket');
  $(form).submit(function(event) {
    event.preventDefault();

    var formData = $(form).serialize();
    $.ajax({
      type: "POST",
      url: $(form).attr("action"),
      data: formData
    });
  });
});

setInterval(function() {
  App.orders.fetch({reset: true});
  
}, 1000);

var App = new( Backbone.View.extend({
  Models: {},
  Views: {},
  Collections: {},

  start: function(options) {
    var orderTicket = new App.Models.OrderTicket({
      session_ids: options.session_ids  
    });
    var orderTicketView = new App.Views.OrderTicket({model: orderTicket});

    this.orders = new App.Collections.Orders(options.orders);
    var ordersView = new App.Views.OrdersView({collection: this.orders});

    this.$el.append(orderTicketView.render().el);
    this.$el.append(ordersView.render().el);
  },
}))({el: $("app")});

App.Models.Order = Backbone.Model.extend({
  urlRoot: "/orders"
});

App.Collections.Orders = Backbone.Collection.extend({
  url: '/orders',
  comparator: 'id'
});
App.Models.OrderTicket = Backbone.Model.extend({});

App.Views.OrderRowView = Backbone.View.extend({
  tagName: 'tr',
  template: _.template(`
<td><% if(open !== "0"){%><button class="btn btn-danger">Cancel</button><% }%></td>
<td><%= symbol %></td>
<td><%= quantity %></td>
<td><%= account %></td>
<td><%= open %></td>
<td><%= closed %></td>
<td><%= prettySide %></td>
<td><%= prettyOrdType %></td>
<td><%= price %></td>
<td><%= stop_price %></td>
<td><%= avg_px %></td>
<td><%= session_id %></td>
`),

  prettySide: function() {
    var sideEnum = this.model.get("side");
    switch(sideEnum) {
      case "1":
        return "Buy";
      case "2":
        return "Sell";
      case "5":
        return "Sell Short";
      case "6":
        return "Sell Short Exempt";
      case "8":
        return "Cross";
      case "9":
        return "Cross Short";
      case "A":
        return "Cross Short Exempt";
    }

    return sideEnum;
  },
  prettyOrdType: function() {
    var ordTypeEnum = this.model.get("ord_type");
    switch (ordTypeEnum) {
      case "1": return "Market";
      case "2": return "Limit";
      case "3": return "Stop";
      case "4": return "Stop Limit";
    };

    return ordTypeEnum;
  },

  render: function() {
    var attribs = _.clone(this.model.attributes);
    attribs.prettySide = this.prettySide();
    attribs.prettyOrdType = this.prettyOrdType();
    this.$el.html(this.template(attribs));
    return this;
  },
  events: {
    "click button": "cancel"
  },
  cancel: function(e) {
    this.model.destroy();
  }
});

App.Views.OrdersView = Backbone.View.extend({
  initialize: function() {
    this.listenTo(this.collection, 'reset', this.addAll);
  },

  render: function() {
    this.$el.html(`
<table class='table table-striped' id='orders'>
  <thead>
    <tr>
      <th></th>
      <th>Symbol</th>
      <th>Quantity</th>
      <th>Account</th>
      <th>Open</th>
      <th>Executed</th>
      <th>Side</th>
      <th>Type</th>
      <th>Limit</th>
      <th>Stop</th>
      <th>AvgPx</th>
      <th>Session</th>
    </tr>
  </thead>
  <tbody>
  </tbody>
</table>`);

    this.collection.forEach(this.addOne, this);
    return this;
  },

  addAll: function() {
    this.$("tbody").empty();
    this.collection.forEach(this.addOne, this);
    return this;
  },

  addOne: function(order) {
    var row = new App.Views.OrderRowView({model: order});
    this.$("tbody").append(row.render().el);
  }
});

App.Views.OrderTicket = Backbone.View.extend({
  template: _.template(`
<form class='form-inline' action='/order' method='POST' id='order-ticket'>
  <p>
    <div class='form-group'>
      <label for='side'>Side</label>
      <select class='form-control' name='side'>
        <option value='1'>Buy</option>
        <option value='2'>Sell</option>
        <option value='5'>Sell Short</option>
        <option value='6'>Sell Short Exempt</option>
        <option value='8'>Cross</option>
        <option value='9'>Cross Short</option>
        <option value='A'>Cross Short Exempt</option>
      </select>
    </div>

    <div class='form-group'>
      <label for='quantity'>Quantity</label>
      <input type='number' class='form-control' name='quantity' placeholder='Quantity' required>
    </div>

    <div class='form-group'>
      <label for='symbol'>Symbol</label>
      <input type='text' class='form-control' name='symbol' placeholder='Symbol' required>
    </div>
  </p>
  <p>
    <div class='form-group'>
      <label for='ordType'>Type</label>
      <select class='form-control' name='ordType' id="ordType">
        <option value='1'>Market</option>
        <option value='2'>Limit</option>
        <option value='3'>Stop</option>
        <option value='4'>Stop Limit</option>
      </select>
    </div>

    <div class='form-group'>
      <label for='limit'>Limit</label>
      <input type='number' step='.01' class='form-control' id="limit" placeholder='Limit' name='price' disabled>
    </div>

    <div class='form-group'>
      <label for='stop'>Stop</label>
      <input type='number' step='.01' class='form-control' id="stop" placeholder='Stop' name='stopPrice' disabled>
    </div>
  </p>

  <p>
    <div class='form-group'>
      <label for='account'>Account</label>
      <input type='text' class='form-control' placeholder='Account' name='account' required>
    </div>

    <div class='form-group'>
      <label for='tif'>TIF</label>
      <select class='form-control' name='tif'>
        <option value='0'>Day</option>
        <option value='3'>IOC</option>
        <option value='2'>OPG</option>
        <option value='1'>GTC</option>
        <option value='5'>GTX</option>
      </select>
    </div>
  </p>

  <p>
    <div class='form-group'>
      <label for='session'>Session</label>
      <select class='form-control' name='session'>
        <% _.each(session_ids, function(i){ %><option><%= i %></option><% }); %>
      </select>
    </div>
  </p>
  <button type='submit' class='btn btn-default'>Submit</button>
</form>
`),
  render: function() {
    this.$el.html(this.template(this.model.attributes));
    return this;
  },

  events: {
    "change #ordType": "updateOrdType",
    submit: "submit"
  },

  submit: function(e) {
    e.preventDefault();
    var order = new App.Models.Order();
    order.set({
      side:       this.$('select[name=side]').val(),
      quantity:   this.$('input[name=quantity]').val(),
      symbol:     this.$('input[name=symbol]').val(),
      ord_type:   this.$('select[name=ordType]').val(),
      price:      this.$('input[name=price]').val(),
      stop_price: this.$('input[name=stopPrice]').val(),
      account:    this.$('input[name=account]').val(),
      tif:        this.$('select[name=tif]').val(),
      session_id: this.$('select[name=session]').val()
    });

    order.save();
  },

  updateOrdType: function() {
    switch(this.$("#ordType option:selected").text()) {
      case "Limit":
        this.$("#limit").prop("disabled", false);
        this.$("#limit").prop("required", true);
        this.$("#stop").prop("disabled", true);
        this.$("#stop").prop("required", false);
      break;

      case "Stop":
        this.$("#limit").prop("disabled", true);
        this.$("#limit").prop("required", false);
        this.$("#stop").prop("disabled", false);
        this.$("#stop").prop("required", true);
      break;

      case "Stop Limit":
        this.$("#limit").prop("disabled", false);
        this.$("#limit").prop("required", true);
        this.$("#stop").prop("disabled", false);
        this.$("#stop").prop("required", true);
      break;

      default:
        this.$("#limit").prop("disabled", true);
        this.$("#stop").prop("disabled", true);
        this.$("#limit").prop("required", false);
        this.$("#stop").prop("required", false);
    }
  }
});


