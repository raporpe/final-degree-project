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
        this.setState({ dropdown: !this.state.dropdown });
    }

    render() {
        let drop = null

        if (this.state.dropdown) {
            drop =
                <div className="zone zone-dropdown">
                    sdfasf
                </div>
        }

        return (
            <div>

                <nav className="zone zone-main" onClick={() => this.toggleDropdown()}>
                    <div>
                        <div>
                            {this.props.currentZone} <i className="fa-solid fa-caret-down"></i>
                        </div>
                    </div>

                {drop}
                </nav>
            </div>
        )
    }
}

export default ZoneSelector;