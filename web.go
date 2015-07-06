package main

const mainpage = `<html><!--Landing page for "/" route -->
<script src="//code.jquery.com/jquery-1.11.3.min.js"></script>
<link rel="stylesheet" href="https://maxcdn.bootstrapcdn.com/bootstrap/3.3.5/css/bootstrap.min.css">

<head>
    <meta content="text/html;charset=utf-8" http-equiv="Content-Type">
    <title>Shorten!</title>

</head>
<script type="text/javascript">
var shorten = function(url) {
    $.ajax({
        url: "http://nindalf.com/shorten",
        method: "POST",
        data: {
            longurl: url
        }
    }).done(function(data) {
        console.log(data);
        $('#result')[0].innerHTML = data.shorturl;
        $('#result').slideDown("slow");
        $('#result').click(function() {
            window.open(data.shorturl, '_blank');
        });
    });
}

var shortenUrl = function() {
    var url = $("#url").val();
    shorten(url);
    return false;
}

$(function() {
    $('#shortenButton').click(function() {
        shortenUrl();
    });
    $('#result').hide();
});
</script>

<body style="font-family: 'Lato', sans-serif; position: fixed; top: 40%; left: 30%;">
    <div class="col-lg-6" style="width:500px;">
        <div class="input-group">
            <form onsubmit="return shortenUrl()">
                <input id="url" type="text" class="form-control" placeholder="Enter long url...">
            </form>
            <span class="input-group-btn">
        <button id="shortenButton" class="btn btn-primary" type="submit">Shorten!</button>

      </span>
        </div>

        <p id="result" style="margin-top:5px; padding:5px; visibility:visible; text-align:center; cursor:pointer;" class="bg-success"></p>
    </div>
</body>

</html>
`
