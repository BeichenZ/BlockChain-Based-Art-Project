import React, { Component } from 'react';
import SVG from './components/SVG'
import Head from 'next/head';



class BlockartSVG extends Component {
  constructor(props) {
    super(props);

    this.state = {
      paths: []
    };
  }

  componentDidMount(){
    setInterval(this.periodicFetchSVG, 100);
  }

  periodicFetchSVG = () => {
    fetch("http://localhost:8080/getshapes", {
      method: 'GET'
    })
    .then(res => res.json())
    .then(response => {
      // console.log(response)
      this.setState({paths: response.SVGs})
    })
    .catch(error => console.error('Error:', error))
  }

  addShapes = () => {
    fetch("http://localhost:5000/addshape", {
      method: 'POST',
      headers: {
        Accept: 'application/json',
        'Content-Type': 'application/json',
      },
      body: JSON.stringify({
        shape: 'circle',
      })
    })
  }

  renderSVG = () => {
    const svgs = this.state.paths.map((svg, index) => {
      return (
        <span>
          <SVG id={index} d={svg.Path} fill={svg.Fill} stroke={svg.Stroke}/>
        </span>
      )
    });
    return svgs
  }


  render() {
    return (
      <div>
       <Head>
         <title>My styled page</title>
         <link href="./css/flex.css" rel="stylesheet" />
       </Head>

         <p>
           CPSC 416
         </p>
         <div>
           <button onClick={this.addShapes.bind(this)}></button>
           <div>
             {this.renderSVG()}
           </div>
         </div>
     </div>

    )
  }
}

export default BlockartSVG
