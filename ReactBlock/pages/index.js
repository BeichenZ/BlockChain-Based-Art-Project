import React, { Component } from 'react';
import SVG from './components/SVG'

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
    fetch("http://localhost:5000/getshapes", {
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
    const svgs = this.state.paths.map((svg) => {
      return (
        <SVG d={svg.Path} fill={svg.Fill} stroke={svg.Stroke}/>

      )
    });
    return svgs
  }


  render() {
    return (
      <div>
        <button onClick={this.addShapes.bind(this)}></button>
        {this.renderSVG()}
      </div>

    )
  }
}

export default BlockartSVG
