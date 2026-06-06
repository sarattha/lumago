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

bool clipRayAxis(float origin, float direction, float minValue, float maxValue, inout float tMin, inout float tMax) {
    if (abs(direction) < 0.00001) {
        return origin >= minValue && origin <= maxValue;
    }
    float t1 = (minValue - origin) / direction;
    float t2 = (maxValue - origin) / direction;
    if (t1 > t2) {
        float tmp = t1;
        t1 = t2;
        t2 = tmp;
    }
    tMin = max(tMin, t1);
    tMax = min(tMax, t2);
    return tMin <= tMax;
}

float viewportEntryTravel(vec2 origin, vec2 direction, float maxTravel) {
    float tMin = 0.0;
    float tMax = maxTravel;
    if (!clipRayAxis(origin.x, direction.x, 0.0, pushData.framebufferSize.x, tMin, tMax)) {
        return -1.0;
    }
    if (!clipRayAxis(origin.y, direction.y, 0.0, pushData.framebufferSize.y, tMin, tMax)) {
        return -1.0;
    }
    if (tMax < 0.0 || tMin > maxTravel) {
        return -1.0;
    }
    if (tMin > 0.0) {
        tMin += min(pushData.cellSize * 0.25, 1.0);
    }
    return max(tMin, 1.0);
}

float raymarchShadow(vec2 pixel) {
    vec2 ray = pixel - pushData.lightPosition;
    float targetDistance = length(ray);
    if (targetDistance <= 1.0 || targetDistance > pushData.lightRadius) {
        return 1.0;
    }

    vec2 direction = ray / targetDistance;
    float travel = viewportEntryTravel(pushData.lightPosition, direction, targetDistance);
    if (travel < 0.0) {
        return 1.0;
    }
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
