package 2º.EU.Ejercicios;

public class Cono {
    private double radio_base;
    private double altura;
    private double generatriz;

    public Cono(int radio_base, int altura, int generatriz){
        this.radio_base = radio_base;
        this.altura = altura;
        this.generatriz = generatriz;
    }

    public double getRadio(){
        return radio_base;
    }

    public double getAltura(){
        return altura;
    }

    public double getGeneratriz(){
        return generatriz;
    }

    public void setRadio(double radio_base){
        this.radio_base = radio_base;
    }

    public void setAltura(double altura){
        this.altura = altura;
    }

    public void setGeneratriz(double generatriz){
        this.generatriz = generatriz;
    }

    public double areaTotal(double radio_base, double altura, double generatriz){
        double pi = 3.14159;
        double area = (pi * radio_base * generatriz) + (pi * radio_base * radio_base);
        return area;
    }
}
