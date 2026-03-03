public class Modelo{
    private Color color1 = Color.RED;
    private Color color2 = Color.RED;
    private int radio = 50;
    public void almacenarColor1(Color color){
        color1 = color;
    }

    public void almacenarColor2(Color color){
        color2 = color;
    }

    public void almacenarRadio(int r){
        radio = r;
    }

    public Color getColor1(){
        return color1;
    }

    public Color getColor2(){
        return color2;
    }

    public int getRadio(){
        return radio;
    }
}