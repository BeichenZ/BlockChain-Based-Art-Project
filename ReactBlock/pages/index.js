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
    fetch("http://localhost:5000/echo", {
      method: 'GET'
    })
    .then(res => res.json())
    .then(response => {
      // console.log(response)
      this.setState({paths: response.SVGs})
    })
    .catch(error => console.error('Error:', error))
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
        {this.renderSVG()}
      </div>

    )
  }
}

export default BlockartSVG
