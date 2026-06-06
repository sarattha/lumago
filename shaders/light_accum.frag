#version 450

layout(location = 0) in vec2 fragUV;
layout(location = 0) out vec4 outLight;

layout(binding = 0) uniform sampler2D sceneNormal;
layout(binding = 1) uniform LightUniforms {
    vec4 ambient;
    uint lightCount;
} uniforms;

struct PointLight {
    vec4 positionRadius;
    vec4 colorIntensity;
    vec4 falloffShadowAlpha;
};

layout(std430, binding = 2) readonly buffer LightBuffer {
    PointLight lights[];
};

layout(binding = 3) uniform sampler2DArray shadowMaps;

float shadowFactor(PointLight light, uint lightIndex, vec2 pixel) {
    if (light.falloffShadowAlpha.y < 0.5) {
        return 1.0;
    }
    vec2 delta = pixel - light.positionRadius.xy;
    float distanceToLight = length(delta);
    float angle = atan(delta.y, delta.x);
    float u = angle / 6.28318530718 + 0.5;
    float nearest = texture(shadowMaps, vec3(u, 0.5, float(lightIndex))).r;
    return distanceToLight <= nearest * light.positionRadius.z + 1.0 ? 1.0 : 0.0;
}

void main() {
    vec2 pixel = gl_FragCoord.xy;
    vec3 normal = normalize(texture(sceneNormal, fragUV).xyz * 2.0 - 1.0);
    vec3 lighting = uniforms.ambient.rgb;

    for (uint i = 0; i < uniforms.lightCount; i++) {
        PointLight light = lights[i];
        vec2 delta = light.positionRadius.xy - pixel;
        float radius = max(light.positionRadius.z, 0.0001);
        float distanceToLight = length(delta);
        float attenuation = clamp(1.0 - distanceToLight / radius, 0.0, 1.0);
        attenuation = pow(attenuation, max(light.falloffShadowAlpha.x, 1.0));

        vec3 lightDir = normalize(vec3(delta / radius, 1.0));
        float ndotl = max(dot(normal, lightDir), 0.0);
        lighting += light.colorIntensity.rgb * light.colorIntensity.a * attenuation * ndotl * shadowFactor(light, i, pixel);
    }

    outLight = vec4(lighting, 1.0);
}
