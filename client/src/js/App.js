import React, { Component } from 'react';
import { Container, Row, Col } from 'reactstrap';

import logo from '../img/logo.svg';

import 'bootstrap/dist/css/bootstrap.min.css';
import '../css/App.scss';

class App extends Component {
  componentDidMount() {
    //Connecting to SSE server
    var source = new EventSource("sse/dashboard")
    var logg = "";
    var currentmsg = "";

    source.onmessage = function (event){
        var dashboard = event.data;
        //If message changed printing it to console
        if (dashboard != currentmsg){
            console.log("OnMessage called:");
            console.dir(event);
            currentmsg = dashboard;
            logg += currentmsg + "<br/>";
            console.log(logg);
            document.getElementById("console").innerHTML = logg;
        }
    }
  }

  render() {
    return (
      <Container>
        <Row>
          <Col md="6">
            <h1>Video Upload</h1>
            <form class="form-signin" method="post" action="/upload" enctype="multipart/form-data">
              <input type="file" name="myfiles" id="myfiles" multiple="multiple"/>
              <input type="submit" name="submit" value="Submit"/>
            </form>
          </Col>
          <Col md="6">
            <div id="main">
            <p>
              <code id="console">
              {/* {"<21:26:49> Upload successful"}<br/>
              {"<21:26:49> Upload successful"}<br/>
              {"<21:26:49> Upload successful"} */}
              </code>
            </p>
          </div>
          </Col>
        </Row>

      </Container>
    );
  }
}

export default App;
