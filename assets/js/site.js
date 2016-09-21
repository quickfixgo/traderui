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
  App.executions.fetch({reset: true});
  
}, 1000);

var App = new( Backbone.View.extend({
  Models: {},
  Views: {},
  Collections: {},

  events: {
    'click a[data-internal]': function(e) {
      e.preventDefault();
      Backbone.history.navigate(e.target.pathname, {trigger: true});
    }
  },

  start: function(options) {
    this.orderTicket = new App.Models.OrderTicket({
      session_ids: options.session_ids  
    });

    this.securityDefinitionForm = new App.Models.SecurityDefinitionForm({
      session_ids: options.session_ids  
    });

    this.orders = new App.Collections.Orders(options.orders);
    this.executions = new App.Collections.Executions(options.executions);
    this.router = new App.Router();

    Backbone.history.start({pushState: true});
  },

  showOrders: function() {
    var orderTicketView = new App.Views.OrderTicket({model: this.orderTicket});
    var ordersView = new App.Views.OrdersView({collection: this.orders});

    $("#app").html(orderTicketView.render().el);
    $("#app").append(ordersView.render().el);
    $("#nav-order").addClass("active");
    $("#nav-execution").removeClass("active");
    $("#nav-secdef").removeClass("active");
  },

  showExecutions: function() {
    var orderTicketView = new App.Views.OrderTicket({model: this.orderTicket});
    var executionsView = new App.Views.Executions({collection: this.executions});

    $("#app").html(orderTicketView.render().el);
    $("#app").append(executionsView.render().el);
    $("#nav-order").removeClass("active");
    $("#nav-execution").addClass("active");
    $("#nav-secdef").removeClass("active");
  },

  showSecurityDefinitions: function() {
    var secDefReq = new App.Views.SecurityDefinitionRequest({model: this.securityDefinitionForm});
    $("#app").html(secDefReq.render().el);
    $("#nav-order").removeClass("active");
    $("#nav-execution").removeClass("active");
    $("#nav-secdef").addClass("active");
  },

  showOrderDetails: function(id) {
    var order = new App.Models.Order({id: id});
    order.fetch({
      success: function() {
        var orderView = new App.Views.OrderDetails({model: order});
        $("#app").html(orderView.render().el);
      },
      error: function() {
        console.log('Failed to fetch!');
      }
    });
  },
  showExecutionDetails: function(id) {
    var execution = new App.Models.Execution({id: id});
    execution.fetch({
      success: function() {
        var executionView = new App.Views.ExecutionDetails({model: execution});
        $("#app").html(executionView.render().el);
      },
      error: function() {
        console.log('Failed to fetch!');
      }
    });
  }
}))({el: document.body});

App.Router = Backbone.Router.extend({
  routes: {
    "": "index", 
    "orders": "index",
    "executions": "executions",
    "secdefs": "secdefs",
    "orders/:id": "orderDetails",
    "executions/:id": "executionDetails",
  },

  index: function(){
    App.showOrders();
  },

  executions: function() {
    App.showExecutions();
  },

  secdefs: function() {
    App.showSecurityDefinitions();
  },

  orderDetails: function(id) {
    App.showOrderDetails(id)
  },

  executionDetails: function(id) {
    App.showExecutionDetails(id)
  }
});

App.Models.Order = Backbone.Model.extend({
  urlRoot: "/orders",
});

App.Models.Execution = Backbone.Model.extend({
  urlRoot: "/executions"
});

App.Models.SecurityDefinitionRequest = Backbone.Model.extend({
  urlRoot: "securitydefinitionrequest"
});

App.Models.OrderTicket = Backbone.Model.extend({});
App.Models.SecurityDefinitionForm = Backbone.Model.extend({});

App.Collections.Orders = Backbone.Collection.extend({
  url: '/orders',
  comparator: 'id'
});

App.Collections.Executions = Backbone.Collection.extend({
  url: '/executions',
  comparator: 'id'
});

App.Views.ExecutionDetails = Backbone.View.extend({
  template: _.template(`
<dl class="dl-horizontal">
  <dt>ID</dt><dd><%= id %></dd> 
	<dt>Symbol</dt><dd><%= symbol %></dd>
	<dt>Quantity</dt><dd><%= quantity %></dd>
	<dt>Session</dt><dd><%= session_id %></dd>
  <dt>Side</dt><dd><%= App.prettySide(side) %></dd>
	<dt>Price</dt><dd><%= price %></dd>
</ul>

</div>
  <a href='#' data-internal='true'>Back</a>
</div>
`),
  render: function() {
    this.$el.html(this.template(this.model.attributes));
    return this;
  },
  events: {
    'click a[data-internal]': function(e) {
      e.preventDefault();
      window.history.back();
    }
  }
});



App.Views.OrderDetails = Backbone.View.extend({
  template: _.template(`
<form class="form-horizontal">
  <div class="form-group">
    <label class="col-sm-2 control-label">ID</label>
    <div class="col-sm-10">
      <p class="form-control-static"><%= id %></p>
    </div>
  </div>
  <div class="form-group">
    <label class="col-sm-2 control-label">ClOrID</label>
    <div class="col-sm-10">
      <p class="form-control-static"><%= clord_id %></p>
    </div>
  </div>
  <div class="form-group">
    <label class="col-sm-2 control-label">Symbol</label>
    <div class="col-sm-10">
      <p class="form-control-static"><%= symbol %></p>
    </div>
  </div>
  <div class="form-group">
    <label class="col-sm-2 control-label">Account</label>
    <div class="col-sm-10">
      <p class="form-control-static"><%= account %></p>
    </div>
  </div>
  <div class="form-group">
    <label class="col-sm-2 control-label">Session</label>
    <div class="col-sm-10">
      <p class="form-control-static"><%= session_id %></p>
    </div>
  </div>
  <div class="form-group">
    <label class="col-sm-2 control-label">Side</label>
    <div class="col-sm-10">
      <p class="form-control-static"><%= App.prettySide(side) %></p>
    </div>
  </div>
  <div class="form-group">
    <label class="col-sm-2 control-label">OrdType</label>
    <div class="col-sm-10">
      <p class="form-control-static"><%= App.prettyOrdType(ord_type) %></p>
    </div>
  </div>
  <div class="form-group">
    <label class="col-sm-2 control-label">Closed</label>
    <div class="col-sm-10">
      <p class="form-control-static"><%= closed %></p>
    </div>
  </div>
  <div class="form-group">
    <label class="col-sm-2 control-label">Open</label>
    <div class="col-sm-10">
      <p class="form-control-static"><%= open %></p>
    </div>
  </div>
  <div class="form-group">
    <label class="col-sm-2 control-label">Avg Px</label>
    <div class="col-sm-10">
      <p class="form-control-static"><%= avg_px %></p>
    </div>
  </div>
  <div class="form-group">
    <label class="col-sm-2 control-label">Security Type</label>
    <div class="col-sm-10">
      <p class="form-control-static"><%= security_type %></p>
    </div>
  </div>
  <div class="form-group">
    <label class="col-sm-2 control-label">Maturity Month Year</label>
    <div class="col-sm-10">
      <p class="form-control-static"><%= maturity_month_year %></p>
    </div>
  </div>
  <div class="form-group">
    <label class="col-sm-2 control-label">Maturity Day</label>
    <div class="col-sm-10">
      <p class="form-control-static"><%= maturity_day %></p>
    </div>
  </div>
  <div class="form-group">
    <label class="col-sm-2 control-label">Put or Call</label>
    <div class="col-sm-10">
      <p class="form-control-static"><%= put_or_call %></p>
    </div>
  </div>
  <div class="form-group">
    <label class="col-sm-2 control-label">Strike Price</label>
    <div class="col-sm-10">
      <p class="form-control-static"><%= strike_price %></p>
    </div>
  </div>

  <% if (open == "0") { %>
  <div class="form-group">
    <label class="col-sm-2 control-label">Quantity</label>
    <div class="col-sm-10">
      <p class="form-control-static"><%= quantity %></p>
    </div>
  </div>
  <div class="form-group">
    <label class="col-sm-2 control-label">Price</label>
    <div class="col-sm-10">
      <p class="form-control-static"><%= price %></p>
    </div>
  </div>
  <div class="form-group">
    <label class="col-sm-2 control-label">Stop Price</label>
    <div class="col-sm-10">
      <p class="form-control-static"><%= stop_price %></p>
    </div>
  </div>

  <% } else { %>

  <div class="form-group">
    <label for="quantity" class="col-sm-2 control-label">Quantity</label>
    <div class="col-sm-10">
      <input type="number" class="form-control" id="quantity" placeholder="Quantity" value="<%= quantity %>" required>
    </div>
  </div>
  <% } %>

</form>
</div>
  <button class="btn btn-danger cancel" <% if(open == "0"){%>disabled<% }%>>Cancel</button>
  <button class="btn btn-warning amend" <% if(open == "0"){%>disabled<% }%>>Amend</button>
  <button class="btn btn-info back">Back</button>
</div>
`),
  render: function() {
    this.$el.html(this.template(this.model.attributes));
    return this;
  },
  events: {
    'click .back': function(e) {
      window.history.back();
    },

    'click .cancel': function(e) {
      this.model.destroy({
        success: function() {
          Backbone.history.navigate("/orders", {trigger: true});
        },
        error: function(model, response) {
          console.log('Failed to cancel!');
          console.log(model);
          console.log(response);
        }
      });
    },

    'click .amend': function(e) {
      
    }
  },
});

App.Views.ExecutionRowView = Backbone.View.extend({
  tagName: 'tr',
  template: _.template(`
<td>
<button class="btn btn-info details">Details</button>
</td>
<td><%= symbol %></td>
<td><%= quantity %></td>
<td><%= App.prettySide(side) %></td>
<td><%= price %></td>
<td><%= session_id %></td>
`),


  render: function() {
    this.$el.html(this.template(this.model.attributes));
    return this;
  },
  events: {
    "click .details": "details"
  },
  details: function(e) {
    Backbone.history.navigate("/executions/" + this.model.get("id"), {trigger: true});
  }
});


App.Views.OrderRowView = Backbone.View.extend({
  tagName: 'tr',
  template: _.template(`
<td>
<button class="btn btn-danger cancel" <% if(open == "0"){%>disabled<% }%>>Cancel</button>
<button class="btn btn-info details">Details</button>
</td>
<td><%= symbol %></td>
<td><%= quantity %></td>
<td><%= account %></td>
<td><%= open %></td>
<td><%= closed %></td>
<td><%= App.prettySide(side) %></td>
<td><%= App.prettyOrdType(ord_type) %></td>
<td><%= price %></td>
<td><%= stop_price %></td>
<td><%= avg_px %></td>
<td><%= session_id %></td>
`),

  render: function() {
    this.$el.html(this.template(this.model.attributes));
    return this;
  },
  events: {
    "click .cancel": "cancel",
    "click .details": "details"
  },
  cancel: function(e) {
    this.model.destroy();
  },

  details: function(e) {
    Backbone.history.navigate("/orders/" + this.model.get("id"), {trigger: true});
  }
});

App.Views.Executions = Backbone.View.extend({
  initialize: function() {
    this.listenTo(this.collection, 'reset', this.addAll);
  },

  render: function() {
    this.$el.html(`
<table class='table table-striped' id='executions'>
  <thead>
    <tr>
      <th></th>
      <th>Symbol</th>
      <th>Quantity</th>
      <th>Side</th>
      <th>Price</th>
      <th>Session</th>
    </tr>
  </thead>
  <tbody>
  </tbody>
</table>`);

    this.addAll();

    return this;
  },

  addAll: function() {
    this.$("tbody").empty();
    this.collection.forEach(this.addOne, this);
    return this;
  },

  addOne: function(execution) {
    var row = new App.Views.ExecutionRowView({model: execution});
    this.$("tbody").append(row.render().el);
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

App.Views.SecurityDefinitionRequest = Backbone.View.extend({
  template: _.template(`
<form class='form-inline'>
  <p>
    <div class='form-group'>
      <label for="security_request_type">Security Request Type</label>
      <select class='form-control' name='security_request_type'>
        <option value="0">Security Identity and Specifications</option>
        <option value="1">Security Identity for the Specifications Provided</option>
        <option value="2">List Security Types</option>
        <option value="3">List Securities</option>
      </select>
    </div>
  </p>
  <p>
    <div class='form-group'>
      <label for='security_type'>SecurityType</label>
      <select class='form-control' name='security_type' id='security_type'>
        <option value='CS'>Common Stock</option>
        <option value='FUT'>Future</option>
        <option value='OPT'>Option</option>
      </select>
    </div>

    <div class='form-group'>
      <label for='symbol'>Symbol</label>
      <input type='text' class='form-control' name='symbol' placeholder='Symbol'>
    </div>
  </p>

  <p>
  <div class='form-group'>
    <label for='session'>Session</label>
    <select class='form-control' name='session'>
      <% _.each(session_ids, function(i){ %><option><%= i %></option><% }); %>
    </select>
  </div>

  <button type='submit' class='btn btn-default'>Submit</button>
  </p>
</form>
  `),

  events: {
    submit: "submit"
  },

  submit: function(e) {
    e.preventDefault();
    var req = new App.Models.SecurityDefinitionRequest();
    req.set({
      session_id:             this.$('select[name=session]').val(),
      security_request_type:  parseInt(this.$('select[name=security_request_type]').val()), 
      security_type:          this.$('select[name=security_type]').val(),
      symbol:                 this.$('input[name=symbol]').val(),
    });
    console.log(this.$('select[name=security_type]').val());
    req.save();
  },

  render: function() {
    this.$el.html(this.template(this.model.attributes));
    return this;
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
  </p>

  <p>
    <div class='form-group'>
      <label for='security_type'>SecurityType</label>
      <select class='form-control' name='security_type' id='security_type'>
        <option value='CS'>Common Stock</option>
        <option value='FUT'>Future</option>
        <option value='OPT'>Option</option>
      </select>
    </div>

    <div class='form-group'>
      <label for='symbol'>Symbol</label>
      <input type='text' class='form-control' name='symbol' placeholder='Symbol' required>
    </div>

    <div class='form-group'>
      <label for='maturity_month_year'>Maturity Month Year</label>
      <input type='text' class='form-control' name='maturity_month_year' id='maturity_month_year' placeholder='Maturity Month Year' disabled>
    </div>

    <div class='form-group'>
      <label for='maturity_day'>Maturity Day</label>
      <input type='number' class='form-control' name='maturity_day' id='maturity_day' placeholder='Maturity Day' disabled>
    </div>

    <div class='form-group'>
      <label for='put_or_call'>Put or Call</label>
      <select class='form-control' name='put_or_call' id='put_or_call' disabled>
        <option value=1>Call</option>
        <option value=0>Put</option>
      </select>
    </div>

    <div class='form-group'>
      <label for='strike_price'>Strike Price</label>
      <input type='number' step='.01' class='form-control' name='strike_price' id='strike_price' placeholder='Strike Price' disabled>
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
      <input type='text' class='form-control' placeholder='Account' name='account'>
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
    "change #security_type": "updateSecurityType",
    submit: "submit"
  },

  submit: function(e) {
    e.preventDefault();
    var order = new App.Models.Order();
    order.set({
      side:                 this.$('select[name=side]').val(),
      quantity:             this.$('input[name=quantity]').val(),
      symbol:               this.$('input[name=symbol]').val(),
      ord_type:             this.$('select[name=ordType]').val(),
      price:                this.$('input[name=price]').val(),
      stop_price:           this.$('input[name=stopPrice]').val(),
      account:              this.$('input[name=account]').val(),
      tif:                  this.$('select[name=tif]').val(),
      session_id:           this.$('select[name=session]').val(),
      security_type:        this.$('select[name=security_type]').val(),
      maturity_month_year:  this.$('input[name=maturity_month_year]').val(),
      maturity_day:         parseInt(this.$('input[name=maturity_day]').val()),
      put_or_call:          parseInt(this.$('select[name=put_or_call]').val()),
      strike_price:         this.$('input[name=strike_price]').val(),
    });

    order.save();
  },

  updateSecurityType: function() {
    switch(this.$("#security_type option:selected").text()) {
      case "Common Stock":
        this.$("#maturity_month_year").attr({disabled: true, required: false});
        this.$("#maturity_day").attr({disabled: true});
        this.$("#put_or_call").attr({disabled: true, required: false});
        this.$("#strike_price").attr({disabled: true, required: false});
        break;
      case "Future":
        this.$("#maturity_month_year").attr({disabled: false, required: true});
        this.$("#maturity_day").attr({disabled: false});
        this.$("#put_or_call").attr({disabled: true, required: false});
        this.$("#strike_price").attr({disabled: true, required: false});
        break;
      case "Option":
        this.$("#maturity_month_year").attr({disabled: false, required: true});
        this.$("#maturity_day").attr({disabled: false});
        this.$("#put_or_call").attr({disabled: false, required: true});
        this.$("#strike_price").attr({disabled: false, required: true});
        break;
    }
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

    }
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

App.prettySide = function(sideEnum) {
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
};

App.prettyOrdType = function(ordTypeEnum) {
  switch (ordTypeEnum) {
    case "1": return "Market";
    case "2": return "Limit";
    case "3": return "Stop";
    case "4": return "Stop Limit";
  };

  return ordTypeEnum;
};




