var update_order_type = function() {
switch($("#ordType option:selected").text()) {
  case "Limit":
    $("#limit").prop("disabled", false);
    $("#limit").prop("required", true);
    $("#stop").prop("disabled", true);
    $("#stop").prop("required", false);
    break;

  case "Stop":
    $("#limit").prop("disabled", true);
    $("#limit").prop("required", false);
    $("#stop").prop("disabled", false);
    $("#stop").prop("required", true);
    break;

  case "Stop Limit":
    $("#limit").prop("disabled", false);
    $("#limit").prop("required", true);
    $("#stop").prop("disabled", false);
    $("#stop").prop("required", true);
    break;

  default:
    $("#limit").prop("disabled", true);
    $("#stop").prop("disabled", true);
    $("#limit").prop("required", false);
    $("#stop").prop("required", false);
  }
};

update_order_type();
$("#ordType").change(update_order_type);

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
  $("#orders").load( "/orders", "#orders");
}, 1000);
