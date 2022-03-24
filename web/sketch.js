let data;
let particles = Array()

function setup() {
    let url = "https://tfg-api.raporpe.dev/v1/state";
    httpGet(url, 'json', false, (response) => {
        // when the HTTP request completes, populate the variable that holds the
        // earthquake data used in the visualization.
        data = response

        Object.keys(data).forEach((device) => {
            let date = Object.keys(data[device])[0]
            let macs = Object.keys(data[device][date])
    
            macs.forEach((mac) => {
                let p = new Particle(mac);
                particles.push(p)
                console.log(p)
            })
        });
    
        particles.forEach((p) => {
            p.createParticle()
        })

    });
    console.log(data);
    createCanvas(windowWidth, windowHeight);
    frameRate(30);


}

// this class describes the properties of a single particle.
class Particle {
    // setting the co-ordinates, radius and the
    // speed of a particle in both the co-ordinates axes.
    constructor(mac) {
        this.x = random(0, width);
        this.y = random(0, height);
        this.r = random(1, 8);
        this.xSpeed = random(-2, 2);
        this.ySpeed = random(-1, 1.5);
        this.mac = mac;
    }

    // creation of a particle.
    createParticle() {
        noStroke();
        fill('rgba(200,169,169,0.5)');
        circle(this.x, this.y, this.r);
        text(this.mac, this.x+5, this.y+4);

    }

    // setting the particle in motion.
    moveParticle() {
        if (this.x < 0 || this.x > width)
            this.xSpeed *= -1;
        if (this.y < 0 || this.y > height)
            this.ySpeed *= -1;
        this.x += this.xSpeed;
        this.y += this.ySpeed;
    }

    // this function creates the connections(lines)
    // between particles which are less than a certain distance apart
    joinParticles(particles) {
        particles.forEach(element => {
            let dis = dist(this.x, this.y, element.x, element.y);
            if (dis < 85) {
                stroke('rgba(255,255,255,0.04)');
                line(this.x, this.y, element.x, element.y);
            }
        });
    }
}

let iterator = 0;

function draw() {
    if (!data) {
        console.log("Waiting for data")
        return
    }
    background('#0f0f0f');

    for (let i = 0; i < particles.length; i++) {
        particles[i].moveParticle()
        particles[i].createParticle()
    }


}