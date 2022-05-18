import React from "react";
import './ZoneSelector.css';

class ZoneSelector extends React.Component {
    constructor(props) {
        super(props)
        this.state = {
            dropdown: false,
        }
    }


    toggleDropdown() {
        this.setState({dropdown: !this.state.dropdown});
        console.log("asdfs")
    }

    render() {
        let drop = null

        if (this.state.dropdown) {
            drop = 
            <div>
            sdfasf
            </div>
        }

        return (
            <div className="zone-main">
                <div onClick={() => this.toggleDropdown}>
                    <div>
                        Leganés <i className="fa-solid fa-caret-down"></i>
                    </div>
                </div>
                {drop}

            </div>
        )
    }
}

export default ZoneSelector;