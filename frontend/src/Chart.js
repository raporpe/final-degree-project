import React from 'react';
import './Chart.css'
import { Bar } from "@visx/shape";
import { Group } from "@visx/group";
import { scaleBand, scaleLinear } from "@visx/scale";
import { AxisBottom, AxisTop } from "@visx/axis";


class Chart extends React.Component {


    render() {
        let width = this.props.width;
        let height = this.props.height;
        let verticalMargin = 40;

        let xMax = width;
        let yMax = height - verticalMargin;

        if (this.props.data === undefined) {
            return (
                <div className='chart-error'>No se han podido cargar los datos</div>
            )
        }

        let xScale = scaleBand({
            range: [0, xMax],
            round: true,
            domain: Object.entries(this.props.data).map((d) => d[0]),
            padding: 0.1
        });


        let yScale = scaleLinear({
            range: [yMax, 0],
            round: true,
            domain: [0, Math.max(...Object.entries(this.props.data).map((d) => d[1] * 100))]
        });

        return (
            <svg width={width} height={height}>
                <rect width={width} height={height} fill="url(#blue)" rx={14} />
                <Group top={verticalMargin / 2}>
                    {Object.entries(this.props.data).map((d) => {
                        const date = d[0]
                        const value = d[1]
                        const barWidth = xScale.bandwidth();
                        const barHeight = yMax - (yScale(value * 100) ?? 0);
                        const barX = xScale(date);
                        const barY = yMax - barHeight;
                        const color = Number(date) === new Date().getHours() ?  "rgba(153, 51, 255, 0.8)" : "rgba(80, 171, 255, 1)"
                        return (
                            <Bar
                                key={`bar-${date}`}
                                x={barX}
                                y={barY}
                                rx={6}
                                width={barWidth}
                                height={barHeight}
                                fill={color}
                            />
                        );
                    })}
                    <AxisBottom
                        numTicks={8}
                        top={yMax}
                        scale={xScale}
                        hideTicks={true}
                        hideZero={true}
                        stroke={'#ffffff'}
                        strokeWidth={0}
                        tickFormat={(t) => String(t) + "h"}
                        tickLabelProps={() => ({
                            fill: '#000000',
                            fontSize: 12,
                            textAnchor: 'middle',
                        })}
                    />
                    <AxisTop
                        numTicks={0}
                        top={-5}
                        scale={xScale}
                        hideTicks={true}
                        hideZero={true}
                        strokeDasharray={6}
                        tickLabelProps={() => ({
                            fill: '#ffeb3b',
                            fontSize: 12,
                            textAnchor: 'middle',
                        })}
                    />
                </Group>
            </svg>
        );
    }
}


export default Chart;