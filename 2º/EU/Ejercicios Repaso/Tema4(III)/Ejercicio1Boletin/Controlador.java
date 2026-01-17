public class Controlador implements ActionListener{
    private Vista vista;

    public Controlador(Vista v){
        vista = v;

        vista.getBoton().addActionListener(this);
    }

    @Override
    public void actionPerformed(ActionEvent e){
        vista.appendTextArea(vista.getField().getText());
    }
}