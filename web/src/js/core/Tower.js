export default class Tower {
    constructor() {
        this.group = this.createTower();
        this.initAnimations();
        this.initSettingsListener();
        this.initScrollReactivity();
    }

    createTower() {
        const group = new THREE.Group();
        this.materials = {};

        // --- Materials ---
        this.materials.stone = new THREE.MeshPhysicalMaterial({
            color: 0x050505,
            roughness: 0.9,
            metalness: 0.1,
            clearcoat: 1.0,
            clearcoatRoughness: 0.1
        });

        this.materials.bone = new THREE.MeshPhysicalMaterial({
            color: 0x1a1a1a,
            roughness: 0.8,
            metalness: 0.2
        });

        this.materials.glow = new THREE.MeshBasicMaterial({
            color: 0x00ffff
        });

        // --- BIOMECHANICAL SPINE ---
        const spineGeometry = new THREE.CylinderGeometry(0.5, 1.0, 10, 32);
        const spine = new THREE.Mesh(spineGeometry, this.materials.stone);
        group.add(spine);

        // --- INSTANCED RIBS (Optimized: 1 draw call instead of 20) ---
        const ribGeom = new THREE.TorusGeometry(1.6, 0.04, 8, 64);
        const ribMaterial = this.materials.bone.clone();
        this.ribInstances = new THREE.InstancedMesh(ribGeom, ribMaterial, 20);
        const dummy = new THREE.Object3D();

        for (let i = 0; i < 20; i++) {
            const h = i * 0.5 - 5;
            const radius = 1.6 + Math.sin(i * 0.4) * 0.4;
            dummy.position.set(0, h, 0);
            dummy.rotation.set(Math.PI / 2, 0, 0);
            dummy.scale.set(radius / 1.6, radius / 1.6, 1);
            dummy.updateMatrix();
            this.ribInstances.setMatrixAt(i, dummy.matrix);

            // Glow nodes on even ribs
            if (i % 2 === 0) {
                const nodeGeom = new THREE.IcosahedronGeometry(0.1, 1);
                const node = new THREE.Mesh(nodeGeom, this.materials.glow);
                node.position.set(radius, h, 0);
                group.add(node);
            }
        }
        this.ribInstances.instanceMatrix.needsUpdate = true;
        group.add(this.ribInstances);

        // --- SINUOUS PIPES (External Veins) ---
        const createPipe = (offset) => {
            const pipePoints = [];
            for (let i = 0; i < 50; i++) {
                const angle = i * 0.2 + offset;
                const r = 2.0 + Math.sin(i * 0.1) * 0.2;
                pipePoints.push(new THREE.Vector3(
                    Math.cos(angle) * r,
                    i * 0.2 - 5,
                    Math.sin(angle) * r
                ));
            }
            const pipeCurve = new THREE.CatmullRomCurve3(pipePoints);
            const pipeGeom = new THREE.TubeGeometry(pipeCurve, 64, 0.06, 8, false);
            return new THREE.Mesh(pipeGeom, this.materials.stone);
        };

        for (let j = 0; j < 6; j++) {
            group.add(createPipe((j / 6) * Math.PI * 2));
        }

        // --- TRUE MATHEMATICAL MOBIUS STRIP ---
        const createMobiusGeometry = (radius, width, segmentsU, segmentsV) => {
            const geometry = new THREE.BufferGeometry();
            const vertices = [];
            const indices = [];
            const uvs = [];

            for (let j = 0; j <= segmentsV; j++) {
                const v = (j / segmentsV - 0.5) * width;
                for (let i = 0; i <= segmentsU; i++) {
                    const u = (i / segmentsU) * Math.PI * 2;
                    const x = (radius + v * Math.cos(u / 2)) * Math.cos(u);
                    const y = (radius + v * Math.cos(u / 2)) * Math.sin(u);
                    const z = v * Math.sin(u / 2);
                    vertices.push(x, y, z);
                    uvs.push(i / segmentsU, j / segmentsV);
                }
            }

            for (let j = 0; j < segmentsV; j++) {
                for (let i = 0; i < segmentsU; i++) {
                    const a = i + (segmentsU + 1) * j;
                    const b = i + (segmentsU + 1) * (j + 1);
                    const c = (i + 1) + (segmentsU + 1) * (j + 1);
                    const d = (i + 1) + (segmentsU + 1) * j;
                    indices.push(a, b, d);
                    indices.push(b, c, d);
                }
            }

            geometry.setIndex(indices);
            geometry.setAttribute('position', new THREE.Float32BufferAttribute(vertices, 3));
            geometry.setAttribute('uv', new THREE.Float32BufferAttribute(uvs, 2));
            geometry.computeVertexNormals();
            return geometry;
        };

        this.mobiusMat = new THREE.ShaderMaterial({
            uniforms: {
                uTime: { value: 0 },
                uIntensity: { value: 1.0 },
                colorLow: { value: new THREE.Color(0x050505) },
                colorHigh: { value: new THREE.Color(0xcccccc) }
            },
            vertexShader: `
                varying vec2 vUv;
                void main() {
                    vUv = uv;
                    gl_Position = projectionMatrix * modelViewMatrix * vec4(position, 1.0);
                }
            `,
            fragmentShader: `
                uniform float uTime;
                uniform float uIntensity;
                uniform vec3 colorLow;
                uniform vec3 colorHigh;
                varying vec2 vUv;
                void main() {
                    float pulse = sin(vUv.x * 6.28 + uTime) * 0.5 + 0.5;
                    vec3 finalColor = mix(colorLow, colorHigh, pulse * uIntensity);
                    float edgeGlow = smoothstep(0.45, 0.5, abs(vUv.y - 0.5));
                    finalColor += colorHigh * edgeGlow * 0.3 * uIntensity;
                    gl_FragColor = vec4(finalColor, 0.9);
                }
            `,
            transparent: true,
            side: THREE.DoubleSide
        });

        const mobiusGeom = createMobiusGeometry(6, 1.2, 128, 20);
        this.mobiusStrip = new THREE.Mesh(mobiusGeom, this.mobiusMat);
        this.mobiusStrip.rotation.x = Math.PI / 2;
        group.add(this.mobiusStrip);

        // --- CENTRAL CORE NODES ---
        const coreNodeGeom = new THREE.IcosahedronGeometry(1.2, 0);
        this.coreNodes = [];
        for (let k = 0; k < 3; k++) {
            const coreNode = new THREE.Mesh(coreNodeGeom, this.materials.stone);
            coreNode.position.y = (k - 1) * 3;
            coreNode.rotation.y = k * Math.PI / 3;

            const innerGlow = new THREE.Mesh(
                new THREE.IcosahedronGeometry(1.0, 0),
                new THREE.MeshStandardMaterial({
                    color: 0x8a2be2,
                    emissive: 0x8a2be2,
                    emissiveIntensity: 3
                })
            );
            coreNode.add(innerGlow);
            this.coreNodes.push(coreNode);
            group.add(coreNode);
        }

        // --- PARTICLE DUST (Small floating points around the tower) ---
        const dustCount = 200;
        const dustGeom = new THREE.BufferGeometry();
        const dustPositions = new Float32Array(dustCount * 3);
        const dustSizes = new Float32Array(dustCount);

        for (let i = 0; i < dustCount; i++) {
            const angle = Math.random() * Math.PI * 2;
            const radius = 2 + Math.random() * 5;
            dustPositions[i * 3] = Math.cos(angle) * radius;
            dustPositions[i * 3 + 1] = (Math.random() - 0.5) * 12;
            dustPositions[i * 3 + 2] = Math.sin(angle) * radius;
            dustSizes[i] = Math.random() * 3 + 1;
        }

        dustGeom.setAttribute('position', new THREE.BufferAttribute(dustPositions, 3));
        dustGeom.setAttribute('size', new THREE.BufferAttribute(dustSizes, 1));

        this.dustMat = new THREE.PointsMaterial({
            color: 0x8a2be2,
            size: 0.05,
            transparent: true,
            opacity: 0.4,
            sizeAttenuation: true
        });

        this.dust = new THREE.Points(dustGeom, this.dustMat);
        group.add(this.dust);

        return group;
    }

    initAnimations() {
        gsap.to(this.group.rotation, {
            y: Math.PI * 2,
            duration: 30,
            repeat: -1,
            ease: "none"
        });

        // Mobius Shader Time and Rotation
        const mobius = this.group.children.find(c => c.material && c.material.uniforms);
        if (mobius) {
            gsap.to(mobius.material.uniforms.uTime, {
                value: 6.28,
                duration: 4,
                repeat: -1,
                ease: "none"
            });
            gsap.to(mobius.rotation, {
                z: Math.PI * 2,
                duration: 20,
                repeat: -1,
                ease: "none"
            });
        }

        // Core node breathing
        this.coreNodes.forEach((node, i) => {
            gsap.to(node.scale, {
                x: 1.1, y: 1.1, z: 1.1,
                duration: 2,
                repeat: -1,
                yoyo: true,
                ease: "sine.inOut",
                delay: i * 0.5
            });
        });

        // Dust gentle rotation
        if (this.dust) {
            gsap.to(this.dust.rotation, {
                y: Math.PI * 2,
                duration: 40,
                repeat: -1,
                ease: "none"
            });
        }
    }

    initScrollReactivity() {
        if (typeof ScrollTrigger === 'undefined') return;

        // Increase glow intensity as user scrolls down
        ScrollTrigger.create({
            trigger: '#main-content',
            start: 'top top',
            end: 'bottom bottom',
            onUpdate: (self) => {
                const progress = self.progress;
                if (this.mobiusMat) {
                    this.mobiusMat.uniforms.uIntensity.value = 0.5 + progress * 1.5;
                }
                if (this.dustMat) {
                    this.dustMat.opacity = 0.2 + progress * 0.5;
                }
            }
        });
    }

    initSettingsListener() {
        window.addEventListener('sovereign_settings_updated', (e) => {
            const state = e.detail;
            const palette = state.visuals.palette;

            let color = 0xcccccc;
            if (palette === 'spectral') color = 0x8a2be2;
            if (palette === 'obsidian') color = 0x111111;

            if (this.mobiusMat) {
                this.mobiusMat.uniforms.colorHigh.value.setHex(color);
            }

            this.group.traverse(obj => {
                if (obj.isMesh && obj.material.emissive) {
                    obj.material.color.setHex(color);
                    obj.material.emissive.setHex(color);
                }
            });
        });
    }

    getMesh() {
        return this.group;
    }
}
