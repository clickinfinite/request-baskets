package main

const (
	BASKET_HTML = `<!DOCTYPE html>
<html>
<head lang="en">
  <title>Request Basket: {{.}}</title>
  <link rel="stylesheet" href="https://maxcdn.bootstrapcdn.com/bootstrap/3.3.6/css/bootstrap.min.css">
  <link rel="stylesheet" href="https://maxcdn.bootstrapcdn.com/bootstrap/3.3.6/css/bootstrap-theme.min.css">
  <script src="https://code.jquery.com/jquery-2.1.4.min.js"></script>
  <script src="https://maxcdn.bootstrapcdn.com/bootstrap/3.3.6/js/bootstrap.min.js"></script>

  <style>
    body { padding-top: 70px; }
    h1 { margin-top: 2px; }
    #more { margin-left: 100px; }
  </style>

  <script>
  (function($) {
    var fetchedCount = 0;
    var totalCount = 0;
    var currentConfig;

    var autoRefresh = false;
    var autoRefreshId;

    function getToken() {
      var token = sessionStorage.getItem("token_{{.}}");
      if (!token) { // fall back to master token if provided
        token = sessionStorage.getItem("master_token");
      }
      return token;
    }

    function onAjaxError(jqXHR) {
      if (jqXHR.status == 401) {
        enableAutoRefresh(false);
        $("#token_dialog").modal({ keyboard : false });
      } else {
        $("#error_message_label").html("HTTP " + jqXHR.status + " - " + jqXHR.statusText);
        $("#error_message_text").html(jqXHR.responseText);
        $("#error_message").modal();
      }
    }

    function escapeHTML(value) {
      return value.replace(/&/g,"&amp;").replace(/</g,"&lt;").replace(/>/g,"&gt;").replace(/"/g,"&quot;");
    }

    function renderRequest(id, request) {
      var path = request.path;
      if (request.query) {
        path += "?";
        if (request.query.length > 70) {
          path += request.query.substring(0, 67) + "...";
        } else {
          path += request.query;
        }
      }

      var headers = [];
      Object.keys(request.headers).map(function(header) {
        headers.push(header + ": " + request.headers[header].join(","));
      });

      var headerClass = "default";
      switch(request.method) {
        case "GET":
          headerClass = "success";
          break;
        case "PUT":
          headerClass = "info";
          break;
        case "POST":
          headerClass = "primary";
          break;
        case "DELETE":
          headerClass = "danger";
          break;
      }

			var html = '<div class="row"><div class="col-md-1"><h4 class="text-' + headerClass +
				'" title="' + new Date(request.date).toString() + '">[' + request.method +
        ']</h4></div><div class="col-md-11"><div class="panel-group" id="' + id + '">' +
        '<div class="panel panel-' + headerClass + '"><div class="panel-heading"><h4 class="panel-title">' + path + '</h4></div></div>' +
        '<div class="panel panel-default"><div class="panel-heading"><h4 class="panel-title">' +
        '<a class="collapsed" data-toggle="collapse" data-parent="#' + id + '" href="#' + id + '_headers">Headers</a></h4></div>' +
        '<div id="' + id + '_headers" class="panel-collapse collapse">' +
        '<div class="panel-body"><pre>' + headers.join('\n') + '</pre></div></div></div>';

      if (request.query) {
        html += '<div class="panel panel-default"><div class="panel-heading"><h4 class="panel-title">' +
          '<a class="collapsed" data-toggle="collapse" data-parent="#' + id + '" href="#' + id + '_query">Query Params</a></h4></div>' +
          '<div id="' + id + '_query" class="panel-collapse collapse">' +
          '<div class="panel-body"><pre>' + request.query.split('&').join('\n') + '</pre></div></div></div>';
      }

      if (request.body) {
        html += '<div class="panel panel-default"><div class="panel-heading"><h4 class="panel-title">' +
          '<a class="collapsed" data-toggle="collapse" data-parent="#' + id + '" href="#' + id + '_body">Body</a></h4></div>' +
          '<div id="' + id + '_body" class="panel-collapse collapse in">' +
          '<div class="panel-body"><pre>' + escapeHTML(request.body) + '</pre></div></div></div>';
      }

      html += '</div></div></div><hr/>';

      return html;
    }

    function addRequests(data) {
      totalCount = data.total_count;
      $("#requests_count").html(data.count + " (" + totalCount + ")");
      if (data.count > 0) {
        $("#empty_basket").addClass("hide");
      } else {
        $("#empty_basket").removeClass("hide");
      }

      if (data && data.requests) {
        var requests = $("#requests");
        var index;
        for (index = 0; index < data.requests.length; ++index) {
          requests.append(renderRequest("req" + fetchedCount, data.requests[index]));
          fetchedCount++;
        }
      }

      if (data.has_more) {
        $("#more").removeClass("hide");
        $("#more_count").html(data.count - fetchedCount);
      } else {
        $("#more").addClass("hide");
        $("#more_count").html("");
      }
    }

    function fetchRequests() {
      $.ajax({
        method: "GET",
        url: "/baskets/{{.}}/requests?skip=" + fetchedCount,
        headers: {
          "Authorization" : getToken()
        }
      }).done(addRequests).fail(onAjaxError);
    }

    function fetchTotalCount() {
      $.ajax({
        method: "GET",
        url: "/baskets/{{.}}/requests?max=0",
        headers: {
          "Authorization" : getToken()
        }
      }).done(function(data) {
        if (data && (data.total_count != totalCount)) {
          refresh();
        }
      }).fail(onAjaxError);
    }

    function updateConfig() {
      if (currentConfig && (
        currentConfig.forward_url != $("#basket_forward_url").val() ||
        currentConfig.capacity != $("#basket_capacity").val()
      )) {
        currentConfig.forward_url = $("#basket_forward_url").val();
        currentConfig.capacity = parseInt($("#basket_capacity").val());

        $.ajax({
          method: "PUT",
          url: "/baskets/{{.}}",
          dataType: "json",
          data: JSON.stringify(currentConfig),
          headers: {
            "Authorization" : getToken()
          }
        }).done(function(data) {
          alert("Basket is reconfigured");
        }).fail(onAjaxError);
      }
    }

    function refresh() {
      $("#requests").html(""); // reset
      fetchedCount = 0;
      fetchRequests(); // fetch latest
    }

    function enableAutoRefresh(enable) {
      if (autoRefresh != enable) {
        var btn = $("#auto_refresh");
        if (enable) {
          autoRefreshId = setInterval(fetchTotalCount, 3000);
          btn.removeClass("btn-default");
          btn.addClass("btn-success");
          btn.attr("title", "Auto-Refresh is Enabled");
        } else {
          clearInterval(autoRefreshId);
          btn.removeClass("btn-success");
          btn.addClass("btn-default");
          btn.attr("title", "Auto-Refresh is Disabled");
        }
        autoRefresh = enable;
      }
    }

    function config() {
      $.ajax({
        method: "GET",
        url: "/baskets/{{.}}",
        headers: {
          "Authorization" : getToken()
        }
      }).done(function(data) {
        if (data) {
          currentConfig = data;
          $("#basket_forward_url").val(currentConfig.forward_url);
          $("#basket_capacity").val(currentConfig.capacity);
          $("#config_dialog").modal();
        }
      }).fail(onAjaxError);
    }

    function deleteRequests() {
      $.ajax({
        method: "DELETE",
        url: "/baskets/{{.}}/requests",
        headers: {
          "Authorization" : getToken()
        }
      }).done(function(data) {
        refresh();
      }).fail(onAjaxError);
    }

    function destroyBasket() {
      $("#destroy_dialog").modal("hide");
      enableAutoRefresh(false);

      $.ajax({
        method: "DELETE",
        url: "/baskets/{{.}}",
        headers: {
          "Authorization" : getToken()
        }
      }).done(function(data) {
        sessionStorage.removeItem("token_{{.}}");
        window.location.href = "/web";
      }).fail(onAjaxError);
    }

    // Initialization
    $(document).ready(function() {
      // dialogs
      $("#token_dialog").on("hidden.bs.modal", function (event) {
        sessionStorage.setItem("token_{{.}}", $("#basket_token").val());
        fetchRequests();
      });
      $("#config_form").on("submit", function(event) {
        $("#config_dialog").modal("hide");
        updateConfig();
        event.preventDefault();
      });
      // buttons
      $("#refresh").on("click", function(event) {
        refresh();
      });
      $("#auto_refresh").on("click", function(event) {
        enableAutoRefresh(!autoRefresh);
      });
      $("#config").on("click", function(event) {
        config();
      });
      $("#delete").on("click", function(event) {
        deleteRequests();
      });
      $("#destroy").on("click", function(event) {
        $("#destroy_dialog").modal();
      });
      $("#destroy_confirmed").on("click", function(event) {
        destroyBasket();
      });
      $("#fetch_more").on("click", function(event) {
        fetchRequests();
      });
      // autorefresh and initial fetch
      if (getToken()) {
        enableAutoRefresh(true);
      }
      fetchRequests();
    });
  })(jQuery);
  </script>
</head>
<body>
  <!-- Fixed navbar -->
  <nav class="navbar navbar-default navbar-fixed-top">
    <div class="container">
      <div class="navbar-header">
        <a class="navbar-brand" href="/web">Request Baskets</a>
      </div>
      <div class="collapse navbar-collapse">
        <form class="navbar-form navbar-right">
          <button id="refresh" type="button" title="Refresh" class="btn btn-default">
            <span class="glyphicon glyphicon-refresh"></span>
          </button>
          <button id="auto_refresh" type="button" title="Auto Refresh" class="btn btn-default">
            <span class="glyphicon glyphicon-repeat"></span>
          </button>
          &nbsp;
          <button id="config" type="button" title="Configure" class="btn btn-default">
            <span class="glyphicon glyphicon-cog"></span>
          </button>
          &nbsp;
          <button id="delete" type="button" title="Delete Requests" class="btn btn-warning">
            <span class="glyphicon glyphicon-fire"></span>
          </button>
          <button id="destroy" type="button" title="Destroy Basket" class="btn btn-danger">
            <span class="glyphicon glyphicon-trash"></span>
          </button>
        </form>
      </div>
    </div>
  </nav>

  <!-- Error message -->
  <div class="modal fade" id="error_message" tabindex="-1">
    <div class="modal-dialog">
      <div class="modal-content panel-danger">
        <div class="modal-header panel-heading">
          <h4 class="modal-title" id="error_message_label">HTTP error</h4>
        </div>
        <div class="modal-body">
          <p id="error_message_text"></p>
        </div>
        <div class="modal-footer">
          <button type="button" class="btn btn-default" data-dismiss="modal">Close</button>
        </div>
      </div>
    </div>
  </div>

  <!-- Login dialog -->
  <form>
  <div class="modal fade" id="token_dialog" tabindex="-1">
    <div class="modal-dialog">
      <div class="modal-content panel-warning">
        <div class="modal-header panel-heading">
          <h4 class="modal-title">Token requred</h4>
        </div>
        <div class="modal-body">
          <p>You are not authorized to access this basket. Please enter this basket token, or choose another basket.</p>
          <div class="form-group">
            <label for="basket_token" class="control-label">Token:</label>
            <input type="password" class="form-control" id="basket_token">
          </div>
        </div>
        <div class="modal-footer">
          <a href="/web" class="btn btn-default">Show Baskets</a>
          <button type="submit" class="btn btn-success" data-dismiss="modal">Authorize</button>
        </div>
      </div>
    </div>
  </div>
  </form>

  <!-- Config dialog -->
  <form id="config_form">
  <div class="modal fade" id="config_dialog" tabindex="-1">
    <div class="modal-dialog">
      <div class="modal-content panel-default">
        <div class="modal-header panel-heading">
          <button type="button" class="close" data-dismiss="modal">&times;</button>
          <h4 class="modal-title" id="config_dialog_label">Configuration</h4>
        </div>
        <div class="modal-body">
          <div class="form-group">
            <label for="basket_forward_url" class="control-label">Forward URL:</label>
            <input type="input" class="form-control" id="basket_forward_url">
            <label for="basket_capacity" class="control-label">Basket Capacity:</label>
            <input type="input" class="form-control" id="basket_capacity">
          </div>
        </div>
        <div class="modal-footer">
          <button type="button" class="btn btn-default" data-dismiss="modal">Cancel</button>
          <button type="submit" class="btn btn-primary">Apply</button>
        </div>
      </div>
    </div>
  </div>
  </form>

  <!-- Destroy dialog -->
  <div class="modal fade" id="destroy_dialog" tabindex="-1">
    <div class="modal-dialog">
      <div class="modal-content panel-danger">
        <div class="modal-header panel-heading">
          <button type="button" class="close" data-dismiss="modal">&times;</button>
          <h4 class="modal-title">Destroy This Basket</h4>
        </div>
        <div class="modal-body">
          <p>Are you sure you want to <strong>permanently destroy</strong> this basket and delete all collected requests?</p>
        </div>
        <div class="modal-footer">
          <button type="button" class="btn btn-default" data-dismiss="modal">Cancel</button>
          <button type="button" class="btn btn-danger" id="destroy_confirmed">Destroy</button>
        </div>
      </div>
    </div>
  </div>

  <div class="container">
    <div class="row">
      <div class="col-md-8">
        <h1>Basket: {{.}}</h1>
      </div>
      <div class="col-md-3 col-md-offset-1">
        <h4><abbr title="Current requests count (Total count)">Requests</abbr>: <span id="requests_count"></span></h4>
      </div>
    </div>
    <hr/>
    <div id="requests">
    </div>
    <div id="more" class="hide">
      <button id="fetch_more" type="button" class="btn btn-default">
        More <span id="more_count" class="badge"></span>
      </button>
    </div>

    <!-- Empty basket -->
    <div class="jumbotron text-center hide" id="empty_basket">
      <h1>Empty basket!</h1>
      <p>This basket is empty, send requests to <kbd>/{{.}}</kbd> and they will appear here.</p>
    </div>
  </div>

  <p>&nbsp;</p>
</body>
</html>`
)