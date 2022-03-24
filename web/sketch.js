function setup() {

    let url = "https://api.raporpe.dev/v1/state"
    httpGet(url, 'json', false, function(response) {
        // when the HTTP request completes, populate the variable that holds the
        // earthquake data used in the visualization.
        data = response;
        console.log(data);
    });

}

function draw() {

}