export default class Engine {
    constructor() {
        this.container = document.getElementById('canvas-container');
        if (!this.container) return;

        this.width = this.container.offsetWidth;
        this.height = this.container.offsetHeight;

        this.scene = new THREE.Scene();
        this.camera = new THREE.PerspectiveCamera(75, this.width / this.height, 0.1, 1000);
        this.camera.position.z = 10;

        // Adaptive quality: detect device capability
        this.quality = this.detectQuality();

        this.renderer = new THREE.WebGLRenderer({
            antialias: this.quality !== 'low',
            alpha: true,
            powerPreference: 'high-performance'
        });
        this.renderer.setSize(this.width, this.height);
        this.renderer.setPixelRatio(Math.min(window.devicePixelRatio, this.quality === 'low' ? 1 : 2));
        this.container.appendChild(this.renderer.domElement);

        this.clock = new THREE.Clock();
        this.objects = [];
        this.frameCount = 0;
        this.lastFpsCheck = performance.now();
        this.currentFps = 60;

        this.initLights();
        this.initLenis();
        this.addEventListeners();
        this.render();
    }

    detectQuality() {
        const canvas = document.createElement('canvas');
        const gl = canvas.getContext('webgl');
        if (!gl) return 'low';
        const debugInfo = gl.getExtension('WEBGL_debug_renderer_info');
        const renderer = debugInfo ? gl.getParameter(debugInfo.UNMASKED_RENDERER_WEBGL) : '';
        // Integrated GPUs get medium, dedicated get high, fallback low
        if (/nvidia|radeon|geforce|rtx|gtx/i.test(renderer)) return 'high';
        if (/intel|adreno|mali/i.test(renderer)) return 'medium';
        return 'medium';
    }

    initLights() {
        const ambientLight = new THREE.AmbientLight(0xffffff, 1.5);
        this.scene.add(ambientLight);

        const pointLight = new THREE.PointLight(0xffffff, 2);
        pointLight.position.set(5, 5, 5);
        this.scene.add(pointLight);

        const directionalLight = new THREE.DirectionalLight(0xffffff, 1);
        directionalLight.position.set(-5, 5, 5);
        this.scene.add(directionalLight);
    }

    initLenis() {
        if (typeof Lenis === 'undefined') return;

        this.lenis = new Lenis({
            duration: 1.2,
            easing: (t) => Math.min(1, 1.001 - Math.pow(2, -10 * t)),
            touchMultiplier: 2
        });

        this.lenis.on('scroll', ScrollTrigger.update);

        gsap.ticker.add((time) => {
            this.lenis.raf(time * 1000);
        });

        gsap.ticker.lagSmoothing(0);
    }

    addEventListeners() {
        window.addEventListener('resize', () => this.onResize());
    }

    onResize() {
        this.width = this.container.offsetWidth;
        this.height = this.container.offsetHeight;
        if (this.width === 0 || this.height === 0) return;

        this.camera.aspect = this.width / this.height;
        this.camera.updateProjectionMatrix();
        this.renderer.setSize(this.width, this.height);
    }

    add(object) {
        this.scene.add(object);
        this.objects.push(object);
    }

    render() {
        const elapsedTime = this.clock.getElapsedTime();

        // Subtle camera float
        this.camera.position.y = Math.sin(elapsedTime * 0.5) * 0.2;

        this.renderer.render(this.scene, this.camera);

        // Adaptive quality: monitor FPS
        this.frameCount++;
        const now = performance.now();
        if (now - this.lastFpsCheck >= 2000) {
            this.currentFps = (this.frameCount / ((now - this.lastFpsCheck) / 1000));
            this.frameCount = 0;
            this.lastFpsCheck = now;

            // Drop pixel ratio if FPS is too low
            if (this.currentFps < 25 && this.renderer.getPixelRatio() > 1) {
                this.renderer.setPixelRatio(1);
            }
        }

        requestAnimationFrame(() => this.render());
    }
}
