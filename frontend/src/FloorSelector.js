import './FloorSelector.css';
import React from 'react';

const pre_floors = new Map()
pre_floors.set(3, "Salas estudio")
pre_floors.set(2, "Hemeroteca")
pre_floors.set(1, "Biblioteca")
pre_floors.set(0, "Planta baja")
pre_floors.set(-1, "Sótano")

let maxKey = (m) => {
  let max = -Infinity
  m.forEach((v, k) => {
    if (k > max) {
      max = k
    }
  })
  return max
}

let minKey = (m) => {
  let min = Infinity
  m.forEach((v, k) => {
    if (k < min) {
      min = k
    }
  })
  return min
}

class FloorSelector extends React.Component {
  constructor(props) {
    super(props)
    this.state = {
      floors: props.floors,
      topFloor: maxKey(props.floors),
      bottomFloor: minKey(props.floors),
    }
  }

  render() {
    let elements = []
    this.props.floors.forEach((v, k) => {
      let styles = ["selector"]

      if (k === this.props.currentFloor) {
        styles.push("selector-active")
      } else {
        styles.push("selector-inactive")
      }

      if (k === this.state.topFloor) {
        styles.push("selector-top")
      }

      if (k === this.state.bottomFloor) {
        styles.push("selector-bottom")
      }

      elements.push(<div className={styles.join(" ")} onClick={() => this.props.onFloorChange(k)} key={k}>{k+" "+v}</div>)
    })

    return (
      <div className="selector-main">
        {elements}
      </div>
    )
  }

}

export default FloorSelector;
