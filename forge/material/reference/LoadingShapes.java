/*
 * Copyright 2024 The Android Open Source Project
 * SPDX-License-Identifier: Apache-2.0
 *
 * MaterialShapes / LoadingIndicator geometry adapted for a browser asset exporter.
 * See docs/reference/components.md for the source and regeneration command.
 */
import androidx.graphics.shapes.*;
import java.lang.reflect.Proxy;
import java.util.*;

class LoadingShapes {
    record Vertex(float x, float y, float radius) {}
    static final RoundedPolygon.Companion C = RoundedPolygon.Companion;

    static RoundedPolygon transform(RoundedPolygon polygon, double angle, double sx, double sy) {
        double a = Math.toRadians(angle);
        PointTransformer mapper = (PointTransformer) Proxy.newProxyInstance(
            PointTransformer.class.getClassLoader(), new Class[]{PointTransformer.class}, (proxy, method, args) -> {
                double x = (float) args[0] * sx, y = (float) args[1] * sy;
                float tx = (float) (x * Math.cos(a) - y * Math.sin(a));
                float ty = (float) (x * Math.sin(a) + y * Math.cos(a));
                return ((long) Float.floatToRawIntBits(tx) << 32) | (Float.floatToRawIntBits(ty) & 0xffffffffL);
            });
        return polygon.transformed(mapper);
    }
    static RoundedPolygon custom(Vertex[] points, int repeats, boolean mirror) {
        List<Vertex> result = new ArrayList<>();
        int sections = mirror ? repeats * 2 : repeats;
        double step = Math.PI * 2 / sections;
        double firstAngle = Math.atan2(points[0].y - .5, points[0].x - .5);
        for (int section = 0; section < sections; section++) {
            for (int index = 0; index < points.length; index++) {
                boolean odd = section % 2 != 0;
                int i = mirror && odd ? points.length - 1 - index : index;
                if (mirror && odd && i == 0) continue;
                Vertex point = points[i];
                double x = point.x - .5, y = point.y - .5;
                double a = Math.atan2(y, x);
                a = section * step + (mirror && odd ? step - a + 2 * firstAngle : a);
                double d = Math.hypot(x, y);
                result.add(new Vertex((float) (.5 + Math.cos(a) * d), (float) (.5 + Math.sin(a) * d), point.radius));
            }
        }
        float[] vertices = new float[result.size() * 2];
        List<CornerRounding> rounds = new ArrayList<>();
        for (int i = 0; i < result.size(); i++) {
            Vertex v = result.get(i); vertices[i * 2] = v.x; vertices[i * 2 + 1] = v.y;
            rounds.add(new CornerRounding(v.radius, 0));
        }
        return RoundedPolygonKt.RoundedPolygon(vertices, new CornerRounding(), rounds, .5f, .5f).normalized();
    }
    static String cubics(List<Cubic> curves) {
        StringJoiner result = new StringJoiner(",", "[", "]");
        for (Cubic c : curves) {
            float[] coordinates = { c.getAnchor0X(), c.getAnchor0Y(), c.getControl0X(), c.getControl0Y(), c.getControl1X(), c.getControl1Y(), c.getAnchor1X(), c.getAnchor1Y() };
            StringJoiner curve = new StringJoiner(",", "[", "]");
            for (float coordinate : coordinates) curve.add(String.format(Locale.ROOT, "%.7f", coordinate));
            result.add(curve.toString());
        }
        return result.toString();
    }
    static float scale(List<RoundedPolygon> polygons) {
        float result = 1;
        for (RoundedPolygon polygon : polygons) {
            float[] bounds = polygon.calculateBounds(), max = polygon.calculateMaxBounds(new float[4]);
            result = Math.min(result, Math.max((bounds[2] - bounds[0]) / (max[2] - max[0]), (bounds[3] - bounds[1]) / (max[3] - max[1])));
        }
        return result;
    }
    static String initial(Morph morph, float progress, float scale) {
        List<Cubic> curves = morph.asCubics(progress);
        // Match Path.getBounds: include every cubic control point.
        double minX=Double.POSITIVE_INFINITY,minY=minX,maxX=-minX,maxY=-minX;
        for(Cubic c:curves) {
            float[] xs={c.getAnchor0X(),c.getControl0X(),c.getControl1X(),c.getAnchor1X()};
            float[] ys={c.getAnchor0Y(),c.getControl0Y(),c.getControl1Y(),c.getAnchor1Y()};
            for(int i=0;i<4;i++){minX=Math.min(minX,xs[i]);maxX=Math.max(maxX,xs[i]);minY=Math.min(minY,ys[i]);maxY=Math.max(maxY,ys[i]);}
        }
        double cx=(minX+maxX)/2,cy=(minY+maxY)/2;
        StringBuilder result=new StringBuilder();
        for(int i=0;i<curves.size();i++) {
            Cubic c=curves.get(i);
            if(i==0)result.append(String.format(Locale.ROOT,"M%.6f,%.6f",(c.getAnchor0X()-cx)*scale+.5,(c.getAnchor0Y()-cy)*scale+.5));
            result.append(String.format(Locale.ROOT,"C%.6f,%.6f %.6f,%.6f %.6f,%.6f",(c.getControl0X()-cx)*scale+.5,(c.getControl0Y()-cy)*scale+.5,(c.getControl1X()-cx)*scale+.5,(c.getControl1Y()-cy)*scale+.5,(c.getAnchor1X()-cx)*scale+.5,(c.getAnchor1Y()-cy)*scale+.5));
        }
        return result+"Z";
    }
    static String sequence(List<RoundedPolygon> polygons, boolean circular) {
        StringJoiner morphs = new StringJoiner(",", "[", "]");
        for (int i = 0; i < polygons.size() - (circular ? 0 : 1); i++) {
            Morph morph = new Morph(polygons.get(i).normalized(), polygons.get((i + 1) % polygons.size()).normalized());
            List<Cubic> start = morph.asCubics(0), end = morph.asCubics(1);
            if (start.size() != end.size()) throw new IllegalStateException("Unmatched morph curves");
            morphs.add("[" + cubics(start) + "," + cubics(end) + "]");
        }
        return "{\"scale\":" + scale(polygons) + ",\"initial\":\"" + initial(new Morph(polygons.get(0).normalized(),polygons.get(1).normalized()),circular?0:.4f,scale(polygons)) + "\",\"morphs\":" + morphs + "}";
    }
    public static void main(String[] args) {
        RoundedPolygon softBurst = custom(new Vertex[]{new Vertex(.193f,.277f,.053f),new Vertex(.176f,.055f,.053f)},10,false);
        RoundedPolygon cookie9 = transform(ShapesKt.star(C,9,1,.8f,new CornerRounding(.5f,0)), -90,1,1).normalized();
        RoundedPolygon pentagon = custom(new Vertex[]{new Vertex(.500f,-.009f,.172f),new Vertex(1.030f,.365f,.164f),new Vertex(.828f,.970f,.169f)},1,true);
        RoundedPolygon pill = custom(new Vertex[]{new Vertex(.961f,.039f,.426f),new Vertex(1.001f,.428f,0),new Vertex(1.000f,.609f,1)},2,true);
        RoundedPolygon sunny = ShapesKt.star(C,8,1,.8f,new CornerRounding(.15f,0)).normalized();
        RoundedPolygon cookie4 = custom(new Vertex[]{new Vertex(1.237f,1.236f,.258f),new Vertex(.500f,.918f,.233f)},4,false);
        RoundedPolygon oval = transform(ShapesKt.circle(C),-45,1,.64).normalized();
        RoundedPolygon circle = transform(ShapesKt.circle(C,10).normalized(),18,1,1).normalized();
        System.out.println("{\"indeterminate\":" + sequence(List.of(softBurst,cookie9,pentagon,pill,sunny,cookie4,oval),true) + ",\"determinate\":" + sequence(List.of(circle,softBurst),false) + "}");
    }
}
