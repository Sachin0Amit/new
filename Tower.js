
export default class Tower {
    constructor() {
        this.group = this.createTower();
        this.initAnimations();
        this.initSettingsListener();
    }

    createTower() {
        const group = new THREE.Group();

        // --- Shared References ---
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

        // --- THE "RIBS" (Skeletal Rings) ---
        for (let i = 0; i < 20; i++) {
            const h = i * 0.5 - 5;
            const radius = 1.6 + Math.sin(i * 0.4) * 0.4;
            const ribGeom = new THREE.TorusGeometry(radius, 0.04, 16, 100);
            const rib = new THREE.Mesh(ribGeom, this.materials.bone);
            rib.position.y = h;
            rib.rotation.x = Math.PI / 2;
            
            if (i % 2 === 0) {
                const nodeGeom = new THREE.IcosahedronGeometry(0.1, 1);
                const node = new THREE.Mesh(nodeGeom, this.materials.glow);
                node.position.set(radius, 0, 0);
                rib.add(node);
            }
            
            group.add(rib);
        }

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
                colorLow: { value: new THREE.Color(0x050505) }, // Obsidian Black
                colorHigh: { value: new THREE.Color(0xcccccc) } // Grayish-White
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
                uniform vec3 colorLow;
                uniform vec3 colorHigh;
                varying vec2 vUv;
                void main() {
                    float pulse = sin(vUv.x * 6.28 + uTime) * 0.5 + 0.5;
                    vec3 finalColor = mix(colorLow, colorHigh, pulse);
                    float edgeGlow = smoothstep(0.45, 0.5, abs(vUv.y - 0.5));
                    finalColor += colorHigh * edgeGlow * 0.3;
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
        for (let k = 0; k < 3; k++) {
            const coreNode = new THREE.Mesh(coreNodeGeom, this.materials.stone);
            coreNode.position.y = (k - 1) * 3;
            coreNode.rotation.y = k * Math.PI / 3;
            
            const innerGlow = new THREE.Mesh(
                new THREE.IcosahedronGeometry(1.0, 0),
                new THREE.MeshStandardMaterial({
                    color: 0x8a2be2, // Purple Core Glow instead of Cyan
                    emissive: 0x8a2be2,
                    emissiveIntensity: 3
                })
            );
            coreNode.add(innerGlow);
            group.add(coreNode);
        }

        return group;
    }

    initAnimations() {
        gsap.to(this.group.rotation, {
            y: Math.PI * 2,
            duration: 30,
            repeat: -1,
            ease: "none"
        });

        // Update Mobius Shader Time and Rotation
        const mobius = this.group.children.find(c => c.material && c.material.uniforms);
        if (mobius) {
            gsap.to(mobius.material.uniforms.uTime, {
                value: 6.28,
                duration: 4,
                repeat: -1,
                ease: "none"
            });
            // Rotate the mobius strip independently
            gsap.to(mobius.rotation, {
                z: Math.PI * 2,
                duration: 20,
                repeat: -1,
                ease: "none"
            });
        }

        this.group.children.filter(c => c.type === 'Mesh' && c.geometry.type === 'IcosahedronGeometry').forEach((node, i) => {
            gsap.to(node.scale, {
                x: 1.1, y: 1.1, z: 1.1,
                duration: 2,
                repeat: -1,
                yoyo: true,
                ease: "sine.inOut",
                delay: i * 0.5
            });
        });
    }

    initSettingsListener() {
        window.addEventListener('sovereign_settings_updated', (e) => {
            const state = e.detail;
            const palette = state.visuals.palette;
            
            let color = 0xcccccc;
            if (palette === 'spectral') color = 0x8a2be2;
            if (palette === 'obsidian') color = 0x111111;

            // Update Mobius
            if (this.mobiusMat) {
                this.mobiusMat.uniforms.colorHigh.value.setHex(color);
            }

            // Update Core Nodes
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
