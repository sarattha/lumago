#version 450

layout(location = 0) in vec2 fragUV;
layout(location = 0) out vec4 outSDFShadow;

layout(binding = 0) uniform sampler2D sdfTexture;

layout(push_constant) uniform SDFShadowPush {
    vec2 lightPosition;
    float lightRadius;
    float cellSize;
    vec2 framebufferSize;
    float maxDistance;
    float softness;
} pushData;

float sampleSDF(vec2 pixel) {
    vec2 uv = pixel / pushData.framebufferSize;
    return texture(sdfTexture, uv).r * pushData.maxDistance;
}

float raymarchShadow(vec2 pixel) {
    vec2 ray = pixel - pushData.lightPosition;
    float targetDistance = length(ray);
    if (targetDistance <= 1.0 || targetDistance > pushData.lightRadius) {
        return 1.0;
    }

    vec2 direction = ray / targetDistance;
    float travel = 1.0;
    float shadow = 1.0;
    for (int i = 0; i < 64 && travel < targetDistance; i++) {
        vec2 samplePoint = pushData.lightPosition + direction * travel;
        float distanceToGeometry = sampleSDF(samplePoint);
        if (distanceToGeometry <= pushData.cellSize * 0.75) {
            return 0.0;
        }
        shadow = min(shadow, clamp(distanceToGeometry / pushData.softness, 0.0, 1.0));
        travel += max(distanceToGeometry * 0.8, pushData.cellSize * 0.5);
    }
    return clamp(shadow, 0.0, 1.0);
}

void main() {
    vec2 pixel = fragUV * pushData.framebufferSize;
    float shadowFactor = raymarchShadow(pixel);
    outSDFShadow = vec4(shadowFactor, shadowFactor, shadowFactor, 1.0);
}
