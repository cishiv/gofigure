setInterval(function(){
   fetch("http://localhost:8084/getHashInfo", {mode: 'cors'}) // Any output from the script will go to the "result" div
  .then(response => response.json())
  .then(data => {
    document.getElementById("hashes").innerHTML = ""
    for (var i = 0; i < data.length; i++) {
      document.getElementById("hashes").innerHTML += "<li>" + data[i] + "</li>";
    }
  })
}, 2000); // Poll every 1000ms

setInterval(function(){
   fetch("http://localhost:8084/getHistory", {mode: 'cors'}) // Any output from the script will go to the "result" div
  .then(response => response.json())
  .then(data => {
    document.getElementById("buildHistory").innerHTML = ""
    for(var i = 0; i< data.length; i++) {
      if (data[i].outcome == "success") {
        document.getElementById("buildHistory").innerHTML += "<li><ul><li><strong>buildID</strong>: <code>"+data[i].buildID+"</code></li><li><strong>time started</strong>: <code>"+data[i].time+"</code></li><li><strong>job</strong>: <code>"+data[i].action+"</code></li><li><strong>status</strong>:<span class=\"tag is-success\">   "+data[i].outcome+"</span></li></ul></li><br>";
        document.getElementById("activeBuild").innerHTML = "<strong>Active Build</strong>: None"
      } else if (data[i].outcome == "started") {
        document.getElementById("buildHistory").innerHTML += "<li><ul><li><strong>buildID</strong>: <code>"+data[i].buildID+"</code></li><li><strong>time started</strong>: <code>"+data[i].time+"</code></li><li><strong>job</strong>: <code>"+data[i].action+"</code></li><li><strong>status</strong>:<span class=\"tag is-warning\">   "+data[i].outcome+"</span><br></li></ul></li><progress class=\"progress is-small is-primary\" max=\"100\">15%</progress><br>";
        document.getElementById("activeBuild").innerHTML = "<strong>Active Build</strong>: None<ul><li><strong>buildID</strong>: <code>"+data[i].buildID+"</code></li><li><strong>time started</strong>: <code>"+data[i].time+"</code></li><li><strong>job</strong>: <code>"+data[i].action+"</code></li><li><strong>status</strong>:<span class=\"tag is-warning\">   "+data[i].outcome+"</span><br></li></ul><progress class=\"progress is-small is-primary\" max=\"100\">15%</progress><br>";
      } else {
        document.getElementById("buildHistory").innerHTML += "<li><ul><li><strong>buildID</strong>: <code>"+data[i].buildID+"</code></li><li><strong>time started</strong>: <code>"+data[i].time+"</code></li><li><strong>job</strong>: <code>"+data[i].action+"</code></li><li><strong>status</strong>:<span class=\"tag is-danger\">   "+data[i].outcome+"</span></li></ul></li><br>";
        document.getElementById("activeBuild").innerHTML = "<strong>Active Build</strong>: None"
      }
    }
  })
}, 2000); // Poll every 1000ms

document.addEventListener('DOMContentLoaded', function() {
  let cardToggles = document.getElementsByClassName('card-toggle');
  for (let i = 0; i < cardToggles.length; i++) {
    cardToggles[i].addEventListener('click', e => {
      e.currentTarget.parentElement.parentElement.childNodes[3].classList.toggle('is-hidden');
    });
  }
});

function startBuild(id) {
    fetch("http://localhost:8084/startBuild/"+id, {mode: 'cors'})
    .then(response => console.log(response))
}
document.addEventListener('DOMContentLoaded', function() {
  fetch("http://localhost:8084/getProjectName", {mode: 'cors'})
  .then(response => response.text())
  .then(data => {
    console.log(data)
    document.getElementById("customTitle").innerHTML += " - " + data.replace(/['"]+/g, '');
    document.getElementById("activeBuild").innerHTML = "<strong>Active Build</strong>: None"
  })
});

document.addEventListener('DOMContentLoaded', function() {
  fetch("http://localhost:8084/getJobs", {mode: 'cors'})
  .then(response => response.json())
  .then(data => {
    for(var i = 0; i< data.length; i++) {
      document.getElementById("management").innerHTML += "<div><p><strong>Job: </strong><code>"+data[i]+"</code></p>  " + "<button onclick=\"startBuild("+i+")\" class=\"button is-success\">Start Build</button></div><br>"
    }
  })
});
